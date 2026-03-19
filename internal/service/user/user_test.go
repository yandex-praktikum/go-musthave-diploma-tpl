package user

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"gmart/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUser_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockAuthRepoIFace(ctrl)
	mockTokenGen := NewMockTokenGeneratorIFace(ctrl)
	svc := &User{authRepo: mockRepo, tokenGenerator: mockTokenGen}

	ctx := context.Background()
	in := &signUpInput{
		UserAgent:     "test-agent",
		XForwardedFor: "1.1.1.1",
	}
	in.Body.Login = "new_user"
	in.Body.Password = "strong_password"

	uID := domain.UserID(100)
	sID := domain.SessionID(uuid.New())
	token := domain.Token("jwt-token-123")
	ttl := 24 * time.Hour

	t.Run("Success Registration", func(t *testing.T) {
		// Ожидаем один вызов SignUp в репо (который внутри делает транзакцию)
		mockRepo.EXPECT().SignUp(ctx, in.Body.Login, in.Body.Password, in.UserAgent, in.XForwardedFor).
			Return(uID, sID, nil)

		mockTokenGen.EXPECT().Generate(uID).Return(token, nil)
		mockRepo.EXPECT().ExpDuration().Return(ttl)

		res, err := svc.SignUp(ctx, in)

		assert.NoError(t, err)
		assert.Equal(t, "Bearer "+token.String(), res.Authorization)
		assert.Contains(t, res.SetCookie, sID.String())
		assert.Equal(t, http.StatusOK, res.Body.Code)
	})

	t.Run("User Already Exists", func(t *testing.T) {
		mockRepo.EXPECT().SignUp(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(domain.UserID(0), domain.SessionID(uuid.Nil), ErrUserAlreadyExists)

		res, err := svc.SignUp(ctx, in)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.True(t, errors.Is(err, ErrUserAlreadyExists))
	})
}

func TestUser_SignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockAuthRepoIFace(ctrl)
	mockTokenGen := NewMockTokenGeneratorIFace(ctrl)
	svc := &User{authRepo: mockRepo, tokenGenerator: mockTokenGen}

	ctx := context.Background()
	uID := domain.UserID(1)
	sID := domain.SessionID(uuid.New())
	token := domain.Token("jwt-token")
	ttl := time.Hour

	t.Run("Fast Track - Valid SessionID", func(t *testing.T) {
		in := &signInInput{SessionID: sID, UserAgent: "ua", XForwardedFor: "ip"}

		// Только RefreshSession, без проверки пароля
		mockRepo.EXPECT().RefreshSession(ctx, sID, in.UserAgent, in.XForwardedFor).Return(uID, nil)
		mockTokenGen.EXPECT().Generate(uID).Return(token, nil)
		mockRepo.EXPECT().ExpDuration().Return(ttl)

		res, err := svc.SignIn(ctx, in)

		assert.NoError(t, err)
		assert.Equal(t, sID.String(), extractCookieValue(res.SetCookie))
	})

	t.Run("Fallback - Expired SessionID", func(t *testing.T) {
		in := &signInInput{SessionID: sID}
		in.Body.Login = "user"
		in.Body.Password = "pass"

		// 1. Попытка обновить сессию провалена
		mockRepo.EXPECT().RefreshSession(ctx, sID, gomock.Any(), gomock.Any()).
			Return(domain.UserID(0), errors.New("expired"))

		// 2. Должен вызваться SignInSleep (проверка пароля)
		mockRepo.EXPECT().SignInSleep(ctx, in.Body.Login, in.Body.Password, gomock.Any(), gomock.Any()).
			Return(uID, sID, nil)

		mockTokenGen.EXPECT().Generate(uID).Return(token, nil)
		mockRepo.EXPECT().ExpDuration().Return(ttl)

		res, err := svc.SignIn(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("No Cookie - Direct Login", func(t *testing.T) {
		in := &signInInput{SessionID: domain.SessionID(uuid.Nil)}
		in.Body.Login = "user"
		in.Body.Password = "pass"

		// RefreshSession не вызывается вообще
		mockRepo.EXPECT().SignInSleep(ctx, in.Body.Login, in.Body.Password, gomock.Any(), gomock.Any()).
			Return(uID, sID, nil)

		mockTokenGen.EXPECT().Generate(uID).Return(token, nil)
		mockRepo.EXPECT().ExpDuration().Return(ttl)

		res, err := svc.SignIn(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		in := &signInInput{SessionID: domain.SessionID(uuid.Nil)}

		mockRepo.EXPECT().SignInSleep(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(domain.UserID(0), domain.SessionID(uuid.Nil), ErrInvalidPassword)

		res, err := svc.SignIn(ctx, in)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.True(t, errors.Is(err, ErrInvalidPassword))
	})
}

// Вспомогательная функция для парсинга значения из Set-Cookie в тестах
func extractCookieValue(setCookie string) string {
	header := http.Header{}
	header.Add("Set-Cookie", setCookie)
	r := http.Response{Header: header}
	if cookies := r.Cookies(); len(cookies) > 0 {
		return cookies[0].Value
	}
	return ""
}

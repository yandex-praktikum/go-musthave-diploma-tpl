package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterSuccess(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	salt := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}
	var saltFN = func(b []byte) (n int, err error) {
		require.Equal(t, 16, len(b))
		copy(b, salt)
		return 16, nil
	}

	saltB64 := base64.URLEncoding.EncodeToString(salt)

	auth := app.NewAuthFN(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage, saltFN)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))
	hex := h.Sum(nil)
	hexB64 := base64.URLEncoding.EncodeToString(hex)

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, loginData *domain.LoginData) {
		require.Equal(t, login, loginData.Login)
		require.Equal(t, saltB64, loginData.Salt)
		require.Equal(t, hexB64, loginData.Hash)
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.NoError(t, err)
}

func TestRegisterLoginBusy(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	auth := app.NewAuth(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, loginData *domain.LoginData) error {
		return fmt.Errorf("err %w", domain.ErrLoginIsBusy)
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrLoginIsBusy)
}

func TestRegisterAnyError(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)

	auth := app.NewAuth(&config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: "secret",
	}, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	regData := &domain.RegistrationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, loginData *domain.LoginData) error {
		return errors.New("custom error")
	}).Times(1)

	err := auth.Register(ctx, regData)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestAuthentificateSuccess(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)
	tokenSecret := "secret"

	gConf := &config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: tokenSecret,
	}

	auth := app.NewAuth(gConf, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	authData := &domain.AuthentificationData{
		Login:    login,
		Password: pass,
	}

	salt := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}
	saltB64 := base64.URLEncoding.EncodeToString(salt)

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))
	hex := h.Sum(nil)
	hexB64 := base64.URLEncoding.EncodeToString(hex)

	userID := 10

	loginData := &domain.LoginData{
		UserID: userID,
		Login:  login,
		Salt:   saltB64,
		Hash:   hexB64,
	}

	mockStorage.EXPECT().GetUserData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lg string) (*domain.LoginData, error) {
		require.Equal(t, login, lg)
		return loginData, nil
	}).Times(1)

	tokenString, err := auth.Login(ctx, authData)
	require.NoError(t, err)

	claims := &domain.Claims{}

	_, err = jwt.ParseWithClaims(string(tokenString), claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	require.NoError(t, err)

	require.Equal(t, userID, claims.UserID)

	// Убедимся в корректности работы метода авторизации
	aData, err := auth.Authorize(ctx, tokenString)
	require.NoError(t, err)
	require.Equal(t, userID, aData.UserID)

	// Ожидаем когда протухнет jwt
	time.Sleep(gConf.TokenExp + 2*time.Second)

	_, err = jwt.ParseWithClaims(string(tokenString), claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	require.Error(t, err)

	aData, err = auth.Authorize(ctx, tokenString)
	require.ErrorIs(t, err, domain.ErrAuthDataIncorrect)
	require.Nil(t, aData)
}

func TestAuthentificateNotFound1(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)
	tokenSecret := "secret"

	gConf := &config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: tokenSecret,
	}

	auth := app.NewAuth(gConf, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	authData := &domain.AuthentificationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().GetUserData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lg string) (*domain.LoginData, error) {
		require.Equal(t, login, lg)
		return nil, nil
	}).Times(1)

	tokenString, err := auth.Login(ctx, authData)
	require.ErrorIs(t, err, domain.ErrWrongLoginPassword)
	require.Equal(t, domain.TokenString(""), tokenString)
}

func TestAuthentificateInternalError(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)
	tokenSecret := "secret"

	gConf := &config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: tokenSecret,
	}

	auth := app.NewAuth(gConf, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	authData := &domain.AuthentificationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().GetUserData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lg string) (*domain.LoginData, error) {
		require.Equal(t, login, lg)
		return nil, domain.ErrServerInternal
	}).Times(1)

	tokenString, err := auth.Login(ctx, authData)
	require.ErrorIs(t, err, domain.ErrServerInternal)
	require.Equal(t, domain.TokenString(""), tokenString)
}

func TestAuthentificateInternalAnyError(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockAuthStorage(ctrl)
	tokenSecret := "secret"

	gConf := &config.GophermartConfig{
		TokenExp:    time.Second * 10,
		TokenSecret: tokenSecret,
	}

	auth := app.NewAuth(gConf, mockStorage)

	login := "user"
	pass := "!eraasd*{1}"
	authData := &domain.AuthentificationData{
		Login:    login,
		Password: pass,
	}

	mockStorage.EXPECT().GetUserData(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lg string) (*domain.LoginData, error) {
		require.Equal(t, login, lg)
		return nil, fmt.Errorf("any error")
	}).Times(1)

	tokenString, err := auth.Login(ctx, authData)
	require.ErrorIs(t, err, domain.ErrServerInternal)
	require.Equal(t, domain.TokenString(""), tokenString)
}

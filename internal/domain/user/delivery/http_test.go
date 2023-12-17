package delivery_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benderr/gophermart/internal/domain/user"
	"github.com/benderr/gophermart/internal/domain/user/delivery"
	"github.com/benderr/gophermart/internal/domain/user/delivery/mocks"
	"github.com/benderr/gophermart/internal/logger/mock_logger"
	"github.com/go-playground/validator"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func newTestServer(mockUsecase delivery.UserUsecase, mockSession delivery.SessionManager) *httptest.Server {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	mockLogger := mock_logger.New()
	delivery.NewUserHandlers(e.Group(""), mockUsecase, mockSession, mockLogger)

	return httptest.NewServer(e)
}

func newRequest(baseServer string, login, pass string) *resty.Request {
	return resty.New().SetBaseURL(baseServer).R().SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON)
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSession := mocks.NewMockSessionManager(ctrl)

	server := newTestServer(mockUsecase, mockSession)
	defer server.Close()

	t.Run("Login success", func(t *testing.T) {
		login := "login"
		pass := "123"
		userid := "testuserid"
		token := "jwttoken"

		mockUsecase.EXPECT().Login(gomock.Any(), login, pass).Return(&user.User{
			Login: login,
			ID:    userid,
		}, nil)

		mockSession.EXPECT().Create(userid).Return(token, nil)

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/login")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "Bearer "+token, resp.Header().Get("Authorization"))
	})

	t.Run("Bad pass", func(t *testing.T) {
		login := "login"
		pass := "123222"

		mockUsecase.EXPECT().Login(gomock.Any(), login, pass).Return(&user.User{
			Login: "",
			ID:    "",
		}, user.ErrBadPass)

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/login")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
		assert.JSONEq(t, string(resp.Body()), `{"message":"user not found"}`)
	})

	t.Run("Invalid contract", func(t *testing.T) {
		login := "login"
		pass := ""

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/login")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		assert.JSONEq(t, string(resp.Body()), `{"message":"invalid request payload"}`)
	})
}

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSession := mocks.NewMockSessionManager(ctrl)

	server := newTestServer(mockUsecase, mockSession)
	defer server.Close()

	t.Run("Register success", func(t *testing.T) {
		login := "login"
		pass := "123"
		userid := "testuserid"
		token := "jwttoken"

		mockUsecase.EXPECT().Register(gomock.Any(), login, pass).Return(&user.User{
			Login: login,
			ID:    userid,
		}, nil)

		mockSession.EXPECT().Create(userid).Return(token, nil)

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/register")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "Bearer "+token, resp.Header().Get("Authorization"))
	})

	t.Run("Already exist", func(t *testing.T) {
		login := "login"
		pass := "123222"

		mockUsecase.EXPECT().Register(gomock.Any(), login, pass).Return(&user.User{
			Login: "",
			ID:    "",
		}, user.ErrLoginExist)

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/register")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusConflict, resp.StatusCode())
		assert.JSONEq(t, string(resp.Body()), `{"message":"already exist"}`)
	})

	t.Run("Invalid contract", func(t *testing.T) {
		login := ""
		pass := ""

		resp, err := newRequest(server.URL, login, pass).
			SetBody(fmt.Sprintf(`{"login":"%v","password":"%v"}`, login, pass)).
			Post("/api/user/register")

		assert.NoError(t, err, "error making HTTP request")

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		assert.JSONEq(t, string(resp.Body()), `{"message":"invalid request payload"}`)
	})
}

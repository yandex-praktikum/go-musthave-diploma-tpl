package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/benderr/gophermart/internal/domain/user"
	"github.com/benderr/gophermart/internal/httputils"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/labstack/echo/v4"
)

type UserUsecase interface {
	Login(ctx context.Context, login, password string) (*user.User, error)
	Register(ctx context.Context, login, password string) (*user.User, error)
}

type SessionManager interface {
	Create(userId string) (string, error)
}

type userHandler struct {
	session SessionManager
	logger  logger.Logger
	UserUsecase
}

type User struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func NewUserHandlers(e *echo.Group, uu UserUsecase, session SessionManager, logger logger.Logger) {
	h := &userHandler{
		UserUsecase: uu,
		session:     session,
		logger:      logger,
	}

	group := e.Group("/api/user")

	group.POST("/login", h.LoginHandler)
	group.POST("/register", h.RegisterHandler)
}

func (u *userHandler) RegisterHandler(c echo.Context) error {
	var usr User

	if err := c.Bind(&usr); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	if err := c.Validate(usr); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	created, err := u.Register(c.Request().Context(), usr.Login, usr.Password)

	if err != nil {
		if errors.Is(err, user.ErrLoginExist) {
			return c.JSON(http.StatusConflict, httputils.Error("already exist"))
		}
		u.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	sessionKey, err := u.session.Create(created.ID)

	if err != nil {
		u.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	c.Response().Header().Add("Authorization", "Bearer "+sessionKey)

	return c.JSON(http.StatusOK, httputils.Ok())
}

func (u *userHandler) LoginHandler(c echo.Context) error {
	var usr User

	if err := c.Bind(&usr); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	if err := c.Validate(&usr); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	existUser, err := u.Login(c.Request().Context(), usr.Login, usr.Password)

	if err != nil {
		u.logger.Errorln("ERROR", err)

		if errors.Is(err, user.ErrNotFound) || errors.Is(err, user.ErrBadPass) {
			return c.JSON(http.StatusUnauthorized, httputils.Error("user not found"))
		}

		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	sessionKey, err := u.session.Create(existUser.ID)

	if err != nil {
		u.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	c.Response().Header().Add("Authorization", "Bearer "+sessionKey)

	return c.JSON(http.StatusOK, httputils.Ok())
}

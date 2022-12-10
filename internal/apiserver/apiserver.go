package apiserver

import (
	"errors"
	"github.com/gorilla/sessions"
	"github.com/iRootPro/gophermart/internal/entity"
	"github.com/iRootPro/gophermart/internal/store"
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"github.com/iRootPro/gophermart/internal/utils"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
)

const sessionName = "gophermart_session"

var (
	errIncorrectLoginOrPassword = errors.New("incorrect login or password")
	errUserIsNotAuthorized      = errors.New("user is not authorized")
)

type APIServer struct {
	config       *Config
	logger       *logrus.Logger
	router       *echo.Echo
	store        *sqlstore.Store
	sessionStore sessions.Store
}

func NewAPIServer(config *Config, sessionsStore sessions.Store) *APIServer {
	return &APIServer{
		config:       config,
		logger:       logrus.New(),
		router:       echo.New(),
		sessionStore: sessionsStore,
	}
}

func (s *APIServer) Start() error {
	s.logger.Infof("starting api server on %s", s.config.RunAddress)

	s.configureRouter()

	if err := s.configureStore(); err != nil {
		log.Fatal(err)
	}

	if err := s.router.Start(s.config.RunAddress); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *APIServer) configureStore() error {
	s.store = sqlstore.New()
	if err := s.store.Open(s.config.DatabaseURI); err != nil {
		log.Fatal(err)
	}

	if err := s.store.CreateTables(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.POST("/api/user/register", s.handleUserCreate())
	s.router.POST("/api/user/login", s.handleUserLogin())

	//private
	s.router.POST("/api/user/orders", s.authUserMiddleware(s.handleLoadOrders()))
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) handleUserCreate() echo.HandlerFunc {
	type request struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		u := &entity.User{
			Login:    req.Login,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session.Values["user_id"] = u.ID
		if err = s.sessionStore.Save(c.Request(), c.Response(), session); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		u.Sanitize()
		return c.JSON(http.StatusOK, u)
	}
}

func (s *APIServer) handleUserLogin() echo.HandlerFunc {
	type request struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		u, err := s.store.User().FindByLogin(req.Login)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectLoginOrPassword)
		}

		if !u.ComparePassword(req.Password) {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectLoginOrPassword)
		}

		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		session.Values["user_id"] = u.ID
		if err = s.sessionStore.Save(c.Request(), c.Response(), session); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, nil)
	}
}

func (s *APIServer) handleLoadOrders() echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || string(body) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		ok := utils.LuhnCheck(string(body))
		if !ok {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid card number")
		}

		userID := c.Get("user").(*entity.User).ID
		order := &entity.Order{
			UserID:      userID,
			OrderNumber: string(body),
		}
		if err := s.store.Order().Create(order); err != nil {
			if err == store.ErrOrderNumberAlreadyExistInThisUser {
				return echo.NewHTTPError(http.StatusOK, err.Error())
			}

			if err == store.ErrOrderNumberAlreadyExistAnotherUser {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}

			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusAccepted, nil)

	}
}

// Middlwewares
func (s *APIServer) authUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := s.sessionStore.Get(c.Request(), sessionName)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		userID, ok := session.Values["user_id"]
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, errUserIsNotAuthorized)
		}

		u, err := s.store.User().FindById(userID.(int))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errUserIsNotAuthorized)

		}

		c.Set("user", u)
		return next(c)
	}
}

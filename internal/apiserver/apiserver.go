package apiserver

import (
	"errors"
	"github.com/gorilla/sessions"
	"github.com/iRootPro/gophermart/internal/entity"
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

const sessionName = "gophermart_session"

var (
	errIncorrectUsernameOrPassword = errors.New("incorrect username or password")
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
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		u := &entity.User{
			Username: req.Username,
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
		return c.JSON(http.StatusCreated, u)
	}
}

func (s *APIServer) handleUserLogin() echo.HandlerFunc {
	type request struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	return func(c echo.Context) error {
		req := &request{}

		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		u, err := s.store.User().FindByUsername(req.Username)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectUsernameOrPassword)
		}

		if !u.ComparePassword(req.Password) {
			return echo.NewHTTPError(http.StatusUnauthorized, errIncorrectUsernameOrPassword)
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

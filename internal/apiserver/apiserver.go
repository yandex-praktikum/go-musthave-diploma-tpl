package apiserver

import (
	"github.com/iRootPro/gophermart/internal/store"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *echo.Echo
	store  *store.Store
}

func NewAPIServer(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: echo.New(),
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
	s.store = store.New()
	if err := s.store.Open(s.config.DatabaseURI); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s *APIServer) configureRouter() {
	s.router.GET("/hello", s.handleHello)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) handleHello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!!!")
}

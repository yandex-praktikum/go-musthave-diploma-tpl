package apiserver

import (
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
)

type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *echo.Echo
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

	s.router.Start(s.config.RunAddress)

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
	return c.String(http.StatusOK, "Hello, World!")
}

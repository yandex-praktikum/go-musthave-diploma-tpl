package handler

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Service *service.ShortenerService
}

func NewHandler(s *service.ShortenerService) *Handler {
	return &Handler{
		Service: s,
	}
}

type App struct {
	Config   config.Config
	Storage  service.URLstorage
	Logger   *logrus.Logger
	Database *sql.DB
}

func InitRoutes(app App) chi.Router {
	app.validateApp()
	shortenerService := &service.ShortenerService{
		Config:   app.Config,
		Storage:  app.Storage,
		Database: app.Database,
	}
	shortenerService.StartDeleteWorker()

	handler := NewHandler(shortenerService)
	apiRouter := chi.NewRouter()
	apiRouter.Use(CompressMiddleware())
	apiRouter.Use(LoggingMiddleware(app.Logger))
	apiRouter.Use(handler.AuthMiddleware)

	apiRouter.Mount("/", ShortenerRoutes(shortenerService))
	return apiRouter
}

func (app *App) validateApp() {
	if app.Logger == nil {
		app.Logger = logrus.New()
	}
}

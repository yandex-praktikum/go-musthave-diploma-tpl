package app

import (
	"context"
	"fmt"

	"net/http"

	"github.com/brisk84/gofemart/internal/config"
	"github.com/brisk84/gofemart/internal/handler"
	"github.com/brisk84/gofemart/internal/router"
	"github.com/brisk84/gofemart/internal/storage"
	"github.com/brisk84/gofemart/internal/usecase"
	"github.com/brisk84/gofemart/pkg/logger"
	"go.uber.org/zap"
)

type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
}

func New(cfg config.Config) (*App, error) {
	lg, err := logger.New(true)
	if err != nil {
		return nil, err
	}

	stor := storage.New(lg, cfg)
	err = stor.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	err = stor.MigrateDown(context.Background())
	if err != nil {
		return nil, err
	}
	err = stor.MigrateUp(context.Background())
	if err != nil {
		return nil, err
	}

	useCase := usecase.New(lg, stor)
	h := handler.New(lg, useCase)

	srv := &http.Server{
		Handler: router.New(h),
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
	}

	return &App{
		HTTPServer: srv,
		logger:     lg,
	}, nil
}

func (app *App) Run() error {
	app.logger.Info("server started", zap.String("addr", app.HTTPServer.Addr))
	return app.HTTPServer.ListenAndServe()
}

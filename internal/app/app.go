package app

import (
	"context"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/config"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/handler"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/logger"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/router"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/storage"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/worker"
	"go.uber.org/zap"
)

// Run запускает http сервер.
func Run(cfg *config.Config) error {
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return err
	}

	storageResult, err := storage.InitializeStorage(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer storageResult.Close()

	services := service.NewServices(storageResult)

	accrualWorker := worker.NewAccrualWorker(cfg.AccrualSystemAddress, services.Order, cfg.AccrualPollInterval)
	go accrualWorker.Run(context.Background())

	h := handler.New(services, cfg.CookieSecret)
	r := router.SetupRouter(h, cfg)
	logger.Log.Info("Starting HTTP server", zap.String("address", cfg.RunAddress))
	return r.Run(cfg.RunAddress)
}

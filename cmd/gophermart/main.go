package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/api"
	"github.com/A-Kuklin/gophermart/internal/accrual"
	"github.com/A-Kuklin/gophermart/internal/auth"
	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/A-Kuklin/gophermart/internal/storage"
	"github.com/A-Kuklin/gophermart/internal/storage/postgres"
	"github.com/A-Kuklin/gophermart/internal/usecases"
)

func main() {
	serverUpCtx, cancelFunc := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancelFunc()

	cfg := config.LoadConfig()
	log := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &logrus.TextFormatter{},
		Level:     cfg.LogLevel,
	}
	var logger logrus.FieldLogger = log

	strg, err := postgres.NewStorage(cfg.DBdsn)
	if err != nil {
		logger.Errorf("Error while creating PSQL connection: %s", err)
	}

	err = strg.Migrate(cfg, "up")
	if err != nil {
		logger.Errorf("Error while up migration: %s", err)
	}
	logger.Info("Migrations were finished")

	storager := storage.NewStorager(strg.DB)
	uc := usecases.NewUseCases(storager, logger)
	apiServer := api.NewAPI(logger, uc)
	auth.Init(cfg)

	accrualStorage := postgres.NewAccrualPSQL(strg.DB)
	accrualClient := accrual.NewClient(accrualStorage, cfg.AccrualAddress, logger)
	accrualClient.Run(serverUpCtx)

	server := serveAPI(cfg, logger, apiServer.SetUpRoutes(cfg))

	<-serverUpCtx.Done()

	logger.Info("Shutdown server...")

	shutdownServer(server)

	logger.Info("Successful server shutdown")
}

func serveAPI(cfg *config.Config, logger logrus.FieldLogger, h http.Handler) *http.Server {
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: h,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("Server error")
		}
	}()

	logger.Infof("Server has started")

	return server
}

func shutdownServer(srv *http.Server) {
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("Server shutdown error: %s", err)
	}
}

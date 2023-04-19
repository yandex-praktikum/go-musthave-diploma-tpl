package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RedWood011/cmd/gophermart/internal/config"
	"github.com/RedWood011/cmd/gophermart/internal/cron"
	"github.com/RedWood011/cmd/gophermart/internal/database/postgres"
	"github.com/RedWood011/cmd/gophermart/internal/logger"
	"github.com/RedWood011/cmd/gophermart/internal/service"
	"github.com/RedWood011/cmd/gophermart/internal/transport/http"
)

func main() {
	ctx := context.Background()
	log := logger.InitLogger()
	cfg := config.New()
	db, err := postgres.NewDatabase(ctx, cfg.DataBaseURI, cfg.CountRepetitionBD)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = db.Ping(ctx)
	if err != nil {
		log.Info("Failed to ping to database")
		return
	}

	serviceHTTP := service.NewService(nil, cfg, log)
	c, err := cron.NewCron(serviceHTTP, log)
	if err != nil {
		log.Info("failed to init cron")

		return
	}

	serverParam := http.ServerParams{
		Service: serviceHTTP,
		Storage: db,
		Cfg:     cfg,
		Logger:  log,
	}

	server := http.NewServer(serverParam)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		log.Info("Shutting down server...")
		server.Shutdown()
	}()
	go func() {
		c.Run(ctx)
	}()

	server.Listen(":3000")
}

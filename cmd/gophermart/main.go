package main

import (
	"context"
	"encoding/json"
	"fmt"
	golog "log"
	"net/http"

	"go.uber.org/zap"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/config"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/handler"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/repository"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/service"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/worker"
	"github.com/AnxVit/go-musthave-diploma-tpl/migrations"
	"github.com/AnxVit/go-musthave-diploma-tpl/pkg/postgresql"
)

var (
	commandUp       = "up"
	defaultLogLevel = "INFO"
)

func main() {
	cfg := initConfig()

	logger.Initialize(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())

	migrations.Migrate(cfg.DatabaseDSN, commandUp, []string{})

	pool := postgresql.New(ctx, cfg.DatabaseDSN)
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()

	repo := repository.New(pool.Pool())

	service := service.NewService(&cfg.Service, repo)

	handler := handler.NewHandler(&cfg.Handler, service)

	logger.Log.Info(fmt.Sprintf("Listen %s", cfg.Addr))

	err := http.ListenAndServe(cfg.Addr, handler)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Listen address: %v", ""), zap.Error(err))
	}

	if cfg.Worker.Enabled {
		workerAccural := worker.NewWorker(ctx, &cfg.Worker, pool.Pool())
		workerDone := make(chan struct{})
		go func() {
			workerAccural.StartWork(ctx)
			workerDone <- struct{}{}
		}()

		cancel()
		workerAccural.Stop()
		<-workerDone
	}
}

// Инициализация конфига, если конига нет, берутся дефолтные параметры
func initConfig() *config.Config {
	var opt Options
	parseFlag(&opt)

	cfg := &config.Config{
		Handler: handler.Config{
			EncryptionKey: "secret-key",
		},
		Service: service.Config{
			SaltLen:   8,
			KeyLength: 16,
		},
		Addr:        opt.Addr,
		DatabaseDSN: opt.DatabaseDSN,
		Worker: worker.Config{
			AccrualSystemAddress: opt.AccrualSystemAddress,
		},
		LogLevel: defaultLogLevel,
	}
	var err error
	if opt.ConfigPath != "" {
		cfg, err = config.NewConfig(opt.ConfigPath)
	}

	if err != nil {
		golog.Fatalf("Could not get config: %v", err)
	}

	cfgStr, err := json.MarshalIndent(cfg, " ", " ")
	if err != nil {
		golog.Fatal(err)
	}

	golog.Println(string(cfgStr))

	return cfg
}

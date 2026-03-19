package main

import (
	"context"
	"fmt"
	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/serve"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {

	// Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем программу
	if err := run(ctx); err != nil {
		slog.Error("server terminated with error",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

}

func run(ctx context.Context) error {

	// Загрузка .env
	_ = godotenv.Load()

	// Конфигурация
	cmdArgs := os.Args[1:]
	cfg := config.NewConfig()
	cfg, err := cfg.Init(
		config.WithEnv(),
		config.WithCmdArgs(&cmdArgs),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize config object: %w", err)
	}
	slog.Info("configuration initialized",
		slog.String("version", cfg.Version()),
	)

	// Регистратор для метрик
	metricsReg := prometheus.NewRegistry()

	// Инстанс PostgreSQL
	pg, err := pgc.NewPgInstance(
		ctx,
		cfg.DBDSN(),
		metrics.NewForPgInstance(metricsReg),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize the postgres instance object: %w", err)
	}
	defer func() {
		slog.Info("closing database connection")
		pg.Close()
	}()

	// Миграции прерывать нельзя
	if err := pg.RunMigrations(context.Background()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Запуск HTTP сервера
	return serve.Serve(ctx, &serve.Input{
		Cfg:        cfg,
		Pg:         pg,
		MetricsReg: metricsReg,
	})
}

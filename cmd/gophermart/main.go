package main

import (
	"context"
	"fmt"
	"gmart/internal/config"
	repoAuth "gmart/internal/repository/auth"
	"gmart/internal/repository/pgc"
	"gmart/internal/telemetry/metrics"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {

	// Запускаем программу
	if err := run(); err != nil {
		slog.Error("server terminated with error",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

}

func run() error {

	// Global Context
	ctx := context.Background()

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

	// Метрики
	// Если метрики запаникуют, то не сможем стартовать
	// И это нормально
	reg := prometheus.NewRegistry()

	// Инстанс PostgreSQL
	pg, err := pgc.NewPgInstance(
		ctx,
		cfg.DBDSN(),
		metrics.NewForPgInstance(reg),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize the postgres instance object: %w", err)
	}
	defer pg.Close()

	// Миграции
	if err := pg.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Репозиторий Auth
	a := repoAuth.NewAuthRepo(
		pg,
		metrics.NewForAuth(reg),
		cfg.SessTTL(),
	)

	slog.Info("AUTH",
		slog.Any("a", a),
	)

	// // Сервис сокращателя: Работаем со строками, удовлетворяющими формату URL
	// srv := service.NewService(ctx, cfg, repo)
	//
	// // HTTP хэндлеры, связанные вызовами с service
	// h := handlers.NewHandlers(cfg, srv)
	//
	// // Ручки HTTP протокола
	// router := handlers.NewRouter(cfg, h)
	//
	// Запуск HTTP сервера
	slog.Info("Starting server",
		slog.String(
			"ListenAddr",
			cfg.ListenAddr(),
		),
		slog.String(
			"AccrualSystemAddr",
			cfg.AccrualSystemAddr(),
		),
		slog.String(
			"PgInstance",
			pg.String(),
		),
	)
	// return handlers.Serve(cfg, router)

	return nil
}

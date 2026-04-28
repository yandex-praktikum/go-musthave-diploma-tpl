package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/KalessinD/gophermart/internal/config"
	gm "github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/KalessinD/gophermart/internal/logger"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	MigrationsDir = "./migrations/"
)

func main() {
	if err := run(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func runHTTPServer(cfg *config.GophermartConfig, log *zap.Logger) error {
	rootCtx, cancel := context.WithCancel(context.Background())
	notifyCtx, _ := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()

	pgdb, err := databaseWorks(rootCtx, cfg, log)
	if err != nil {
		return err
	}

	router, err := gm.NewRouter(cfg, log, pgdb)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           router,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		BaseContext:       func(net.Listener) context.Context { return notifyCtx },
	}

	// Graceful Shutdown
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		log.Info("Server started at " + cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	return waitServer(rootCtx, server, errCh, cfg, log)
}

func databaseWorks(ctx context.Context, cfg *config.GophermartConfig, log *zap.Logger) (*sql.DB, error) {
	if cfg.PsqlDSN == "" {
		return nil, errors.New("database_dsn is empty")
	}

	pgdb, err := gm.PsqlConnect(ctx, cfg.PsqlDSN, log)
	if err != nil {
		return nil, err
	}

	migrations := []string{"migrations/000001_init_project.up.sql"}
	migrator := gm.NewPgMigrator(pgdb)

	err = migrator.Apply(ctx, MigrationsDir, migrations)
	if err != nil {
		log.Error("Can't apply migration", zap.Error(err))
		return nil, err
	}

	return pgdb, nil
}

func run() error {
	zapLogger, err := logger.NewLogger(config.IsProduction())
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	defer func() { _ = zapLogger.Sync() }()

	cfg, err := config.NewGophermartConfig(flag.CommandLine, os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return runHTTPServer(cfg, zapLogger)
}

func waitServer(serverCtx context.Context, server *http.Server, errCh chan error, cfg *config.GophermartConfig, log *zap.Logger) error {
	// Канал для перехвата системных сигналов
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем либо ошибки запуска, либо сигнала завершения
	select {
	case err := <-errCh:
		return fmt.Errorf("server startup error: %w", err)
	case sig := <-quit:
		log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	}

	// Даем серверу время на завершение текущих запросов
	ctx, cancel := context.WithTimeout(serverCtx, cfg.GracefullShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info("Server stopped gracefully")
	return nil
}

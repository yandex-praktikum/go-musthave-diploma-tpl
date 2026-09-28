package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/accrual"
	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
	"github.com/true18/go-musthave-diploma-tpl/internal/config"
	"github.com/true18/go-musthave-diploma-tpl/internal/handler"
	"github.com/true18/go-musthave-diploma-tpl/internal/migrations"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/server"
	"github.com/true18/go-musthave-diploma-tpl/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DatabaseURI == "" {
		return errors.New("DATABASE_URI is required")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer logger.Sync()

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		return fmt.Errorf("ping database: %w", err)
	}
	cancel()

	if err := migrations.Up(db); err != nil {
		return err
	}

	repo := repository.New(db)
	appService := service.New(repo)

	authManager, err := auth.NewRandom()
	if err != nil {
		return err
	}

	appHandler := handler.New(appService, authManager)
	httpServer := server.New(cfg.RunAddress, appHandler, authManager, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.AccrualSystemAddress == "" {
		logger.Info("accrual system address is not set; worker disabled")
	} else {
		client := accrual.NewClient(cfg.AccrualSystemAddress, nil)
		worker := service.NewWorker(repo, client, logger)
		go worker.Run(ctx)
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting gophermart", zap.String("address", cfg.RunAddress))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaw13/go-musthave-diploma/internal/accrual"
	"github.com/nikitaw13/go-musthave-diploma/internal/auth"
	"github.com/nikitaw13/go-musthave-diploma/internal/handler"
	"github.com/nikitaw13/go-musthave-diploma/internal/logger"
	"github.com/nikitaw13/go-musthave-diploma/internal/service"
	"github.com/nikitaw13/go-musthave-diploma/internal/storage"
	"go.uber.org/zap"
)

func main() {
	parseEnvs()
	parseFlags()

	err := validateArguments()
	if err != nil {
		log.Fatal(err)
	}

	err = run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// The first SIGINT/SIGTERM triggers a graceful shutdown; a second one kills
	// the process immediately (default signal behavior) — for those unwilling
	// to wait for the graceful path to finish.
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Worker context, cancellable on demand: cancelWorker() also stops the worker
	// when the server itself fails — a bare signal context is only cancelled by a signal.
	workerCtx, cancelWorker := context.WithCancel(signalCtx)
	defer cancelWorker()

	err := logger.InitLogger(argumentLogLevel)
	if err != nil {
		return err
	}

	if argumentJWTSecretKey == "" {
		buf := make([]byte, 32)
		_, err := rand.Read(buf)
		if err != nil {
			return fmt.Errorf("generate JWT secret: %w", err)
		}
		argumentJWTSecretKey = base64.StdEncoding.EncodeToString(buf)
		logger.Logger.Warn("JWT secret is not provided, using ephemeral; sessions won't survive restart")
	}

	pool, err := pgxpool.New(context.Background(), argumentDSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	err = pool.Ping(pingCtx)
	if err != nil {
		return fmt.Errorf("database is not reachable: %w", err)
	}

	err = runMigrations(pool)
	if err != nil {
		return fmt.Errorf("database migrations failed: %w", err)
	}

	var (
		databaseTimeouts = []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}
		classifier       = storage.NewPostgresErrorClassifier()
	)

	var (
		pgStorage      = storage.NewPostgresStorage(pool, databaseTimeouts, classifier, logger.Logger)
		jwtManager     = auth.NewJWTManager(argumentJWTSecretKey, logger.Logger)
		loyaltyService = service.NewLoyaltyService(pgStorage, jwtManager, logger.Logger)
		apiHandler     = handler.NewAPIHandler(loyaltyService, logger.Logger)
		router         = apiHandler.Routes()
		srv            = &http.Server{
			Addr:         argumentServiceAddress,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      router,
		}
	)

	// Buffered by 1 so the server goroutine never blocks on send, even if the
	// main goroutine has already left the select via workerCtx.Done().
	serverErr := make(chan error, 1)

	// ListenAndServe runs in its own goroutine, leaving the main goroutine free
	// to call Shutdown on srv once a stop signal arrives.
	go func() {
		logger.Logger.Info("server is running", zap.String("address", argumentServiceAddress), zap.String("log_level", argumentLogLevel))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}

	}()

	var (
		accrualClient = accrual.NewAccrualClient(argumentAccrualAddress, logger.Logger)
		worker        = accrual.NewAccrualWorker(accrualClient, pgStorage, logger.Logger, time.Second*1)
	)

	go worker.ProcessNewOrders(workerCtx)

	var serverFail error
	select {
	case serverFail = <-serverErr:
		// The server died on its own — stop the worker as well.
		cancelWorker()
	case <-workerCtx.Done():
	}

	// Stop capturing signals: a second Ctrl+C must now kill the process
	// instead of being swallowed by the still-active notification.
	stop()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	// Shutdown drains in-flight requests (up to 5s), then we wait for the worker
	// to finish its current poll (up to 10s). The deferred pool.Close() runs only
	// after run() returns, so the worker never hits a closed pool.
	err = srv.Shutdown(shutdownCtx)
	select {
	case <-worker.Done():
	case <-time.After(10 * time.Second):
		logger.Logger.Error("worker didn't stop in time")
	}
	if serverFail != nil {
		return serverFail
	}
	return err
}

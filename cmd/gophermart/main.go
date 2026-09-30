package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaw13/go-musthave-diploma/internal/accrual"
	"github.com/nikitaw13/go-musthave-diploma/internal/auth"
	"github.com/nikitaw13/go-musthave-diploma/internal/handler"
	"github.com/nikitaw13/go-musthave-diploma/internal/logger"
	"github.com/nikitaw13/go-musthave-diploma/internal/service"
	"github.com/nikitaw13/go-musthave-diploma/internal/storage"
)

func main() {
	parseFlags()
	err := parseEnvs()
	if err != nil {
		log.Fatal(err)
	}

	err = validateFlags()
	if err != nil {
		log.Fatal(err)
	}

	err = run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := logger.InitLogger("DEBUG")
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(context.Background(), flagDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(pingCtx)
	if err != nil {
		return fmt.Errorf("database is not reachable: %w", err)
	}

	orderQueue := make(chan string, flagQueueCapacity)
	var timeouts = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	var classifier = storage.NewPostgresErrorClassifier()

	pgStorage := storage.NewPostgresStorage(pool, timeouts, classifier, logger.Logger)
	jwtManager := auth.NewJWTManager(flagJWTSecretKey, logger.Logger)
	loyaltyService := service.NewLoyaltyService(pgStorage, jwtManager, orderQueue, logger.Logger)
	apiHandler := handler.NewAPIHandler(loyaltyService, logger.Logger)
	router := apiHandler.Routes()
	srv := &http.Server{
		Addr:         flagServiceAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}
	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accrualClient := accrual.NewAccrualClient(flagAccrualAddress)
	worker := accrual.NewAccrualWorker(accrualClient, logger.Logger, time.Second*1)
	go worker.Run(shutdownCtx)

	return srv.Shutdown(shutdownCtx)
}

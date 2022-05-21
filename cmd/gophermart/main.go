package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/r4start/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/r4start/go-musthave-diploma-tpl/internal/app"
)

type config struct {
	ServerAddress            string
	AccrualSystemAddress     string
	DatabaseConnectionString string
}

func main() {
	cfg := config{
		ServerAddress: ":8080",
	}

	flag.StringVar(&cfg.ServerAddress, "a", os.Getenv("RUN_ADDRESS"), "")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", os.Getenv("ACCRUAL_SYSTEM_ADDRESS"), "")
	flag.StringVar(&cfg.DatabaseConnectionString, "d", os.Getenv("DATABASE_URI"), "")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("failed to initialize logger: %+v", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if len(cfg.DatabaseConnectionString) == 0 {
		logger.Fatal("Empty database connection string")
	}

	dbConn, err := pgxpool.Connect(context.Background(), cfg.DatabaseConnectionString)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	storageCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userStorage, err := storage.NewUserDatabaseStorage(storageCtx, dbConn)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	marketServer, err := app.NewServer(serverCtx, logger, userStorage)
	if err != nil {
		logger.Fatal("Failed to initialize app server", zap.Error(err))
	}

	server := &http.Server{Addr: cfg.ServerAddress, Handler: marketServer}
	server.ListenAndServe()
}

package gophermart

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/KalessinD/gophermart/internal/clients"
	handler "github.com/KalessinD/gophermart/internal/handlers"
	"github.com/KalessinD/gophermart/internal/processors"
	"github.com/KalessinD/gophermart/internal/repositories/file"
	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"
	service "github.com/KalessinD/gophermart/internal/services"

	"github.com/KalessinD/gophermart/internal/config"
	mw "github.com/KalessinD/gophermart/internal/middleware"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	maxConnectionRetries           = 3
	waitIntervalBetweenConnections = time.Second * 3
)

func PsqlConnect(ctx context.Context, dsn string, log *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Error("Failed to parse DSN", zap.Error(err))
		return nil, fmt.Errorf("parsing DSN: %w", err)
	}

	var lastErr error

	for attempt := range maxConnectionRetries {
		lastErr = db.PingContext(ctx)
		if lastErr == nil {
			log.Info("Successfully connected to PostgreSQL", zap.Int("attempt", attempt))
			break
		}

		log.Warn("Failed to connect to PostgreSQL, retrying...",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxConnectionRetries),
			zap.Duration("interval", waitIntervalBetweenConnections),
			zap.Error(lastErr),
		)

		if attempt < maxConnectionRetries {
			select {
			case <-ctx.Done():
				log.Warn("Database connection canceled by context during retry")
				db.Close()
				return nil, ctx.Err()
			case <-time.After(waitIntervalBetweenConnections):
				// Время вышло, идем на следующий круг
			}
		}
	}

	if lastErr != nil {
		log.Error("Failed to connect to PostgreSQL after all retries", zap.Error(lastErr))
		db.Close()
		return nil, fmt.Errorf("db connection failed after %d retries: %w", maxConnectionRetries, lastErr)
	}

	go func() {
		<-ctx.Done()
		log.Info("Closing database connection due to context cancellation")
		db.Close()
	}()

	return db, nil
}

func GetBaseRouter(cfg *config.GophermartConfig, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(mw.LoggerMiddleware(log))
	router.Use(middleware.Timeout(cfg.ProcessingTimeout))

	return router
}

func NewRouter(ctx context.Context, cfg *config.GophermartConfig, log *zap.Logger, pgdb *sql.DB) (http.Handler, error) {
	router := GetBaseRouter(cfg, log)

	commonUserHandler := handler.NewCommonHandler(
		service.NewCommonService(
			repository.NewSQLStorage(pgdb),
		),
		service.NewAuthService(cfg.EncryptionKey),
	)

	// свободный доступ
	router.Route("/api/user/", func(r chi.Router) {
		r.Post("/login", commonUserHandler.Login)
		r.Post("/register", commonUserHandler.Register)
	})

	linkCh := make(chan *processors.Task, cfg.WorkerPoolChanBuffer)
	pauseCh := make(chan time.Duration, 1)

	orderService := service.NewOrderService(
		repository.NewSQLStorage(pgdb),
		clients.NewAccrualClient(cfg.AccrualAddress, cfg.AccrualClientTImeout, log),
		linkCh,
	)

	ordersHandler := handler.NewOrdersHandler(
		orderService,
	)

	dumper, err := file.NewFileStorage(file.AvroType, cfg.DumperStoragePath)
	if err != nil {
		return nil, err
	}

	processor, err := processors.NewQueueProcessor(
		cfg.QueueWorkers,
		cfg.QueueBufSize,
		log,
		linkCh,
		pauseCh,
		dumper,
		orderService.ProcessAccrualTask,
	)
	if err != nil {
		return nil, err
	}

	// если есть, что восстановить из прошлых задач - восстановим
	processor.RestoreQueue()
	processor.Start(ctx)

	go func() {
		// nolint:revive
		for range ctx.Done() {
		}
		close(linkCh)
	}()

	balancesHandler := handler.NewBalancesHandler(
		service.NewBalanceService(
			repository.NewSQLStorage(pgdb),
		),
	)

	// JWT Auth
	router.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(cfg.EncryptionKey))

		r.Post("/api/user/orders", ordersHandler.AddOrder)
		r.Get("/api/user/orders", ordersHandler.ListOrders)
		r.Get("/api/user/balance", balancesHandler.GetLoyalityBalance)
		r.Post("/api/user/balance/withdraw", balancesHandler.WithdrawBalance)
		r.Get("/api/user/withdrawals", balancesHandler.ListWithdrawals)
	})

	return router, nil
}

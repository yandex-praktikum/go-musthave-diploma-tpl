package gophermart

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	handler "github.com/KalessinD/gophermart/internal/handlers"
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

func NewRouter(cfg *config.GophermartConfig, log *zap.Logger, pgdb *sql.DB) (http.Handler, error) {
	router := GetBaseRouter(cfg, log)

	commonUserHandler := handler.NewCommonHandler(
		service.NewCommonAction(
			repository.NewSQLStorage(pgdb),
		),
		service.NewAuthService(cfg.EncryptionKey),
	)

	// свободный доступ
	router.Route("/api/user/", func(r chi.Router) {
		r.Post("/login", commonUserHandler.Login)
		r.Post("/register", commonUserHandler.Register)
	})

	ordersHandler := handler.NewOrdersHandler(
		service.NewOrderActions(
			repository.NewSQLStorage(pgdb),
		),
	)

	balancesHandler := handler.NewBalancesHandler(
		service.NewBalanceActions(
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

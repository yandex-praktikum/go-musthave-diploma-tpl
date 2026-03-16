package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	accrualSystemHTTP "github.com/kdv2001/loyalty/internal/clients/accrualSystem/http"
	httpHandlers "github.com/kdv2001/loyalty/internal/handlers/http"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/store/postgress/auth"
	"github.com/kdv2001/loyalty/internal/store/postgress/loyalty"
	"github.com/kdv2001/loyalty/internal/store/postgress/session"
	loyaltyUseCase "github.com/kdv2001/loyalty/internal/usecases/loyalty"
	"github.com/kdv2001/loyalty/internal/usecases/user"
)

func main() {
	log.Fatal(initService())
}

func initService() error {
	ctx := context.Background()
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to init looger: %w", err)
	}
	sugarLogger := zapLog.Sugar()
	ctx = logger.ToContext(ctx, sugarLogger)

	initValues, err := initFlags()
	if err != nil {
		return err
	}

	postgresConn, err := pgxpool.New(ctx, initValues.postgresDSN)
	if err != nil {
		return err
	}
	if err = postgresConn.Ping(ctx); err != nil {
		return err
	}

	authStore, err := auth.NewImplementation(ctx, postgresConn)
	if err != nil {
		return err
	}

	sessionStore, err := session.NewImplementation(ctx, postgresConn)
	if err != nil {
		return err
	}

	loyaltyStore, err := loyalty.NewImplementation(ctx, postgresConn)
	if err != nil {
		return err
	}

	u, err := url.Parse(initValues.accrualSystemAddress)
	if err != nil {
		return fmt.Errorf("failed to init looger: %w", err)
	}

	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	accrualClient := accrualSystemHTTP.NewClient(client, *u)
	userUC := user.NewImplementation(authStore, sessionStore)
	loyaltyUC := loyaltyUseCase.NewImplementation(ctx, accrualClient, loyaltyStore)

	chiMux := chi.NewMux()
	chiMux.Use(
		httpHandlers.AddLoggerToContextMiddleware(sugarLogger),
		httpHandlers.ResponseMiddleware(),
		httpHandlers.RequestMiddleware())

	authMW := httpHandlers.NewAuthMiddleware(userUC)

	handlers := httpHandlers.New(userUC, loyaltyUC)
	chiMux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.Register)
		r.Post("/login", handlers.Login)

		rWithAuth := r.With(authMW.Middleware)
		rWithAuth.Post("/orders", handlers.AddOrder)
		rWithAuth.Get("/orders", handlers.GetOrders)
		rWithAuth.Get("/balance", handlers.GetBalance)
		rWithAuth.Post("/balance/withdraw", handlers.WithdrawalPoints)
		rWithAuth.Get("/withdrawals", handlers.GetWithdrawals)
	})

	sugarLogger.Infof("start listen and serve")
	return http.ListenAndServe(initValues.serverAddr, chiMux)
}

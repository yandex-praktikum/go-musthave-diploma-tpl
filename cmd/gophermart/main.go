package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	httpHandlers "github.com/kdv2001/loyalty/internal/handlers/http"
	"github.com/kdv2001/loyalty/internal/store/postgress/auth"
	"github.com/kdv2001/loyalty/internal/store/postgress/session"
	"github.com/kdv2001/loyalty/internal/usecases/user"
)

func main() {
	log.Fatal(initService())
}

func initService() error {
	ctx := context.Background()

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

	userUC := user.NewImplementation(authStore, sessionStore)
	log, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to init looger: %w", err)
	}
	sugarLogger := log.Sugar()

	chiMux := chi.NewMux()
	chiMux.Use(
		httpHandlers.AddLoggerToContextMiddleware(sugarLogger),
		httpHandlers.ResponseMiddleware(),
		httpHandlers.RequestMiddleware())

	authMW := httpHandlers.NewAuthMiddleware(userUC)

	handlers := httpHandlers.New(userUC)
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

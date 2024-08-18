package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/eac0de/gophermart/internal/database"
	"github.com/eac0de/gophermart/internal/handlers"
	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/jwt"
	"github.com/eac0de/gophermart/pkg/logger"
	"github.com/eac0de/gophermart/pkg/middlewares"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
)

const SK = "supersecretkey"

func main() {
	logger.InitLogger("info")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	database, err := database.NewDatabase(ctx, "host=localhost user=postgres password=351762 dbname=gophermart sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer database.Close()
	router := mux.NewRouter()

	router.Use(middlewares.GetLoggerMiddleware())
	router.Use(middlewares.GetGzipMiddleware())
	tokenService := jwt.NewJWTTokenService(SK, 30*time.Minute, 48*time.Hour, database)
	router.Use(middlewares.GetJWTAuthMiddleware(tokenService, database))

	AuthService := services.NewAuthService(SK, database)
	authHandlers := handlers.NewAuthHandlers(tokenService, AuthService)
	router.HandleFunc("/api/user/register", authHandlers.RegisterHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/user/login", authHandlers.LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/user/refresh_token", authHandlers.RefreshTokenHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/user/change_password", authHandlers.ChangePasswordHandler).Methods(http.MethodPatch)

	userService := services.NewUserService(database)
	userHandlers := handlers.NewUserHandlers(userService)
	router.HandleFunc("/api/user/me", userHandlers.UserHandler).Methods(http.MethodGet, http.MethodPatch, http.MethodDelete)

	client := resty.New()
	orderService := services.NewOrderService(database, client)
	go orderService.StartProcessingOrders(ctx)
	orderHandlers := handlers.NewOrderHandlers(orderService)
	router.HandleFunc("/api/user/orders", orderHandlers.OrderHandler).Methods(http.MethodGet, http.MethodPost)

	balanceService := services.NewBalanceService(database)
	balanceHandlers := handlers.NewBalanceHandlers(balanceService)
	router.HandleFunc("/api/user/balance", balanceHandlers.BalanceHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/user/withdrawals", balanceHandlers.GetWithdrawalsHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/user/balance/withdraw", balanceHandlers.CrateWithdrawalsHandler).Methods(http.MethodPost)

	log.Printf("Server http://localhost:8000 is running. Press Ctrl+C to stop")
	http.ListenAndServe(":8000", router)
}

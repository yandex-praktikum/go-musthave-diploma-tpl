package router

import (
	"net/http"

	"github.com/eac0de/gophermart/internal/handlers"
	"github.com/gorilla/mux"
)

type AppRouter struct {
	*mux.Router
}

func NewAppRouter(
	orderHandlers *handlers.OrderHandlers,
	authHandlers *handlers.AuthHandlers,
	balanceHandlers *handlers.BalanceHandlers,
	userHandlers *handlers.UserHandlers,
	middlewares ...mux.MiddlewareFunc,
) *AppRouter {
	appRouter := &AppRouter{mux.NewRouter()}
	for _, middleware := range middlewares {
		appRouter.Use(middleware)
	}
	appRouter.HandleFunc("/api/user/register", authHandlers.RegisterHandler).Methods(http.MethodPost)
	appRouter.HandleFunc("/api/user/login", authHandlers.LoginHandler).Methods(http.MethodPost)
	appRouter.HandleFunc("/api/user/refresh_token", authHandlers.RefreshTokenHandler).Methods(http.MethodPost)
	appRouter.HandleFunc("/api/user/change_password", authHandlers.ChangePasswordHandler).Methods(http.MethodPatch)

	appRouter.HandleFunc("/api/user/me", userHandlers.GetUserHandler).Methods(http.MethodGet)
	appRouter.HandleFunc("/api/user/me", userHandlers.PatchUserHandler).Methods(http.MethodPatch)
	appRouter.HandleFunc("/api/user/me", userHandlers.DeleteUserHandler).Methods(http.MethodDelete)

	appRouter.HandleFunc("/api/user/orders", orderHandlers.GetListOrderHandler).Methods(http.MethodGet)
	appRouter.HandleFunc("/api/user/orders", orderHandlers.PostOrderHandler).Methods(http.MethodPost)

	appRouter.HandleFunc("/api/user/balance", balanceHandlers.BalanceHandler).Methods(http.MethodGet)
	appRouter.HandleFunc("/api/user/withdrawals", balanceHandlers.GetWithdrawalsHandler).Methods(http.MethodGet)
	appRouter.HandleFunc("/api/user/balance/withdraw", balanceHandlers.CreateWithdrawalsHandler).Methods(http.MethodPost)
	return appRouter
}

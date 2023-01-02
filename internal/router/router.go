package router

import (
	"net/http"

	"github.com/brisk84/gofemart/internal/handler"
	"github.com/gorilla/mux"
)

func New(h *handler.Handler) http.Handler {
	r := mux.NewRouter()
	r.Use(SetJSONHeader)

	r.HandleFunc("/healthz", h.HealthCheck).Methods(http.MethodGet)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(SetJSONHeader)

	apiRouter.HandleFunc("/user/register", h.Register).Methods(http.MethodPost)
	apiRouter.HandleFunc("/user/login", h.Login).Methods(http.MethodPost)
	apiRouter.HandleFunc("/user/orders", h.UserOrdersPost).Methods(http.MethodPost)
	apiRouter.HandleFunc("/user/orders", h.UserOrdersGet).Methods(http.MethodGet)
	apiRouter.HandleFunc("/user/balance", h.UserBalanceGet).Methods(http.MethodGet)
	apiRouter.HandleFunc("/user/balance/withdraw", h.UserBalanceWithdraw).Methods(http.MethodPost)
	apiRouter.HandleFunc("/user/withdrawals", h.UserWithdrawals).Methods(http.MethodGet)

	// apiRouter.HandleFunc("/user/orders", h.UserOrders).Methods(http.MethodPost)

	apiRouter.HandleFunc("/users/{id}", h.GetUser).Methods(http.MethodGet)
	apiRouter.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)

	apiRouter.HandleFunc("/orders/{id}", h.GetOrder).Methods(http.MethodGet)
	apiRouter.HandleFunc("/orders", h.CreateOrder).Methods(http.MethodPost)

	return r
}

func SetJSONHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

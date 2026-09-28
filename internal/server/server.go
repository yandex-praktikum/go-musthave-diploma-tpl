// Package server wires HTTP routing and middleware.
package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
	"github.com/true18/go-musthave-diploma-tpl/internal/handler"
	"github.com/true18/go-musthave-diploma-tpl/internal/middleware"
)

// New creates an HTTP server for the Gophermart API.
func New(address string, handler *handler.Handler, authManager *auth.Manager, logger *zap.Logger) *http.Server {
	if logger == nil {
		logger = zap.NewNop()
	}

	mux := http.NewServeMux()
	mux.Handle("/api/user/register", method(http.MethodPost, http.HandlerFunc(handler.Register)))
	mux.Handle("/api/user/login", method(http.MethodPost, http.HandlerFunc(handler.Login)))

	private := middleware.Auth(authManager)
	mux.Handle("/api/user/orders", private(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.UploadOrder(w, r)
		case http.MethodGet:
			handler.Orders(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/user/balance", private(method(http.MethodGet, http.HandlerFunc(handler.Balance))))
	mux.Handle("/api/user/balance/withdraw", private(method(http.MethodPost, http.HandlerFunc(handler.Withdraw))))
	mux.Handle("/api/user/withdrawals", private(method(http.MethodGet, http.HandlerFunc(handler.Withdrawals))))

	root := middleware.Logging(logger)(middleware.Gzip(mux))
	return &http.Server{
		Addr:              address,
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func method(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expected {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

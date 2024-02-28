package router

import (
	"net/http"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/handlers"
	"github.com/Azcarot/GopherMarketProject/internal/middleware"
	"github.com/Azcarot/GopherMarketProject/internal/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var Flag utils.Flags

func MakeRouter(flag utils.Flags) *chi.Mux {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	middleware.Sugar = *logger.Sugar()
	r := chi.NewRouter()
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				handlers.ActualiseOrders(flag, quit)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	r.Use(middleware.WithLogging)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", http.HandlerFunc(handlers.Registration))
		r.With(middleware.CheckAuthorization).Post("/login", http.HandlerFunc(handlers.LoginUser))
		r.With(middleware.CheckAuthorization).Post("/orders", http.HandlerFunc(handlers.Order))
		r.With(middleware.CheckAuthorization).Post("/balance/withdraw", http.HandlerFunc(handlers.Withdraw))
		r.With(middleware.CheckAuthorization).Get("/orders", http.HandlerFunc(handlers.GetOrders))
		r.With(middleware.CheckAuthorization).Get("/balance", http.HandlerFunc(handlers.GetBalance))
		r.With(middleware.CheckAuthorization).Get("/withdrawals", http.HandlerFunc(handlers.GetWithdrawals))
	})
	return r
}

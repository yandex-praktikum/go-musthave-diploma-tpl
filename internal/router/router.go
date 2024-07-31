package router

import (
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
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	// делаем регистратор SugaredLogger
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
		r.Post("/register", handlers.Registration().ServeHTTP)
		r.Post("/login", handlers.LoginUser().ServeHTTP)
		r.Post("/orders", handlers.Order(flag).ServeHTTP)
		r.Post("/balance/withdraw", handlers.Withdraw().ServeHTTP)
		r.Get("/orders", handlers.GetOrders().ServeHTTP)
		r.Get("/balance", handlers.GetBalance().ServeHTTP)
		r.Get("/withdrawals", handlers.GetWithdrawals().ServeHTTP)
	})
	return r
}

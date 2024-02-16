package router

import (
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
	r.Use(middleware.WithLogging)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.Registration().ServeHTTP)
		r.Post("/login", handlers.LoginUser().ServeHTTP)
		// r.Post("/orders", Storagehandler.HandleJSONGetMetrics(flag).ServeHTTP)
		// r.Post("/balance/withdraw", Storagehandler.HandlePostMetrics().ServeHTTP)
		// r.Get("/orders", Storagehandler.HandleGetMetrics().ServeHTTP)
		// r.Get("/balance", storage.CheckDBConnection(storage.DB).ServeHTTP)
		// r.Get("/withdrawals", storage.CheckDBConnection(storage.DB).ServeHTTP)
	})
	return r
}

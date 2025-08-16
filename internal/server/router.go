package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/middleware"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"go.uber.org/zap"
)

// Router настраивает HTTP роутер
type Router struct {
	handlers *Handlers
	router   *chi.Mux
}

// NewRouter создает новый роутер
func NewRouter(storage Storage, authService *services.AuthService, accrualService *services.AccrualService, logger *zap.Logger) *Router {
	handlers := NewHandlers(storage, authService, accrualService, logger)
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.GzipMiddleware)

	// Все маршруты /api/user
	router.Route("/api/user", func(r chi.Router) {
		// Публичные
		r.Post("/register", handlers.RegisterHandler)
		r.Post("/login", handlers.LoginHandler)

		// Защищённые
		r.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware(authService))
			protected.Post("/orders", handlers.UploadOrderHandler)
			protected.Get("/orders", handlers.GetOrdersHandler)
			protected.Get("/balance", handlers.GetBalanceHandler)
			protected.Post("/balance/withdraw", handlers.WithdrawHandler)
			protected.Get("/withdrawals", handlers.GetWithdrawalsHandler)
		})
	})

	return &Router{
		handlers: handlers,
		router:   router,
	}
}

// GetRouter возвращает настроенный роутер
func (r *Router) GetRouter() *chi.Mux {
	return r.router
}

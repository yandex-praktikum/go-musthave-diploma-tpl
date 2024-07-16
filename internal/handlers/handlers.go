package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/adapters/postgres"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type HTTPHandlers struct {
	logger      zap.SugaredLogger
	userService *service.UserService
}

func NewHTTPHandlers(logger zap.SugaredLogger, userService *service.UserService) *HTTPHandlers {
	return &HTTPHandlers{
		logger:      logger,
		userService: userService,
	}
}

func CreateServeMux(logger zap.SugaredLogger, pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	userRepo := postgres.NewPgUserRepository(pool)
	userService := service.NewUserService(userRepo)
	handlers := NewHTTPHandlers(logger, userService)

	r.Post("/api/user/register", WithAuth(false, WithLogging(logger, handlers.registerUser)))

	return r
}

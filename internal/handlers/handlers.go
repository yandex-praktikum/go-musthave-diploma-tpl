package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"net/http"
)

type HTTPHandlers struct {
	logger zap.SugaredLogger
	pool   *pgxpool.Pool
}

func (h HTTPHandlers) registerUser(writer http.ResponseWriter, request *http.Request) {

}

func NewHTTPHandlers(logger zap.SugaredLogger, pool *pgxpool.Pool) *HTTPHandlers {
	return &HTTPHandlers{
		logger: logger,
		pool:   pool,
	}
}

func CreateServeMux(logger zap.SugaredLogger, pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	handlers := NewHTTPHandlers(logger, pool)

	r.Post("/api/user/register", WithAuth(false, WithLogging(logger, handlers.registerUser)))

	return r
}

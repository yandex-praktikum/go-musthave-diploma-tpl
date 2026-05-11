package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router создаёт и настраивает chi роутер с базовыми middleware
func Router() *chi.Mux {
	r := chi.NewMux()

	// Базовые middleware от chi
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	// Наши кастомные middleware
	r.Use(Recovery)
	r.Use(RequestLogger)
	r.Use(Gzip)
	r.Use(GzipRequest)

	return r
}

// MountAuthMiddleware применяет middleware аутентификации к группе роутов
func MountAuthMiddleware(r *chi.Mux, authMiddleware http.Handler, paths ...string) {
	for _, path := range paths {
		r.Mount(path, authMiddleware)
	}
}

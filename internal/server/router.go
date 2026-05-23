package server

import (
	"context"
	"net/http"
)

// Router создаёт и настраивает стандартный ServeMux с middleware
func Router() *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}

// WithMiddleware применяет middleware к http.Handler
func WithMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Применяем middleware в обратном порядке (последний выполняется первым)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// RealIP middleware для определения реального IP клиента
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем X-Forwarded-For и X-Real-IP заголовки
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r = r.WithContext(context.WithValue(r.Context(), "remote_addr", ip))
		} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
			r = r.WithContext(context.WithValue(r.Context(), "remote_addr", ip))
		}
		next.ServeHTTP(w, r)
	})
}

// MountAuthMiddleware применяет middleware аутентификации к группе роутов
// Для стандартного mux просто регистрируем роуты с уже применённым middleware
func MountAuthMiddleware(mux *http.ServeMux, authMiddleware http.Handler, paths ...string) {
	for _, path := range paths {
		mux.Handle(path, authMiddleware)
	}
}

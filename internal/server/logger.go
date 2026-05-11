package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"go-musthave-diploma-tpl/internal/logger"
)

// responseWriter обёртка для захвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// RequestLogger middleware для логгирования запросов
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// Получаем request ID если есть
		reqID := middleware.GetReqID(r.Context())
		if reqID == "" {
			reqID = "-"
		}

		logger.Info("HTTP request",
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"bytes", rw.bytes,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

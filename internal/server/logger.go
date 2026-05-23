package server

import (
	"context"
	"net/http"
	"time"

	"go-musthave-diploma-tpl/internal/logger"
)

// contextKey тип для ключей контекста
type contextKey int

const (
	requestIDKey contextKey = iota
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

// RequestID middleware для генерации request ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			// Генерируем простой ID из timestamp
			reqID = time.Now().Format("20060102150405.000000000")
		}

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestID извлекает request ID из контекста
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return "-"
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

		// Получаем request ID из контекста
		reqID := getRequestID(r.Context())

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

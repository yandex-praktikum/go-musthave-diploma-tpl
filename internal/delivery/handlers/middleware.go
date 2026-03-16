package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// ==================== Константы ====================

const (
	logFieldMethod     = "method"
	logFieldPath       = "path"
	logFieldDuration   = "duration"
	logFieldStatus     = "status"
	logFieldBytes      = "bytes"
	logMessageRequest  = "HTTP request"
	logMessageResponse = "HTTP response"
)

// ==================== Middleware логирования ====================

// loggingMiddleware создает middleware для логирования HTTP запросов и ответов
func (h *UserHandler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обертку для ResponseWriter для перехвата статуса и размера ответа
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Продолжаем обработку запроса
		next.ServeHTTP(ww, r)

		// Вычисляем время выполнения
		duration := time.Since(start)

		// Логируем детали запроса
		h.logRequestDetails(r, duration)

		// Логируем детали ответа
		h.logResponseDetails(ww)
	})
}

// logRequestDetails логирует детали входящего HTTP запроса
func (h *UserHandler) logRequestDetails(r *http.Request, duration time.Duration) {
	h.logger.Info(logMessageRequest,
		zap.String(logFieldMethod, r.Method),
		zap.String(logFieldPath, r.URL.Path),
		zap.Duration(logFieldDuration, duration),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	)
}

// logResponseDetails логирует детали исходящего HTTP ответа
func (h *UserHandler) logResponseDetails(ww middleware.WrapResponseWriter) {
	// Используем соответствующий уровень логирования в зависимости от статуса
	logEntry := h.logger.With(
		zap.Int(logFieldStatus, ww.Status()),
		zap.Int(logFieldBytes, ww.BytesWritten()),
	)

	switch {
	case ww.Status() >= 500:
		logEntry.Error("Server error response")
	case ww.Status() >= 400:
		logEntry.Warn("Client error response")
	case ww.Status() >= 300:
		logEntry.Info("Redirection response")
	default:
		logEntry.Info(logMessageResponse)
	}
}

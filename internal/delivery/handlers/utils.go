package handlers

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

// ==================== Константы ====================

const (
	logFieldComponent      = "component"
	componentRequestBody   = "request_body"
	errorMessageReadBody   = "Failed to read request body"
	errorMessageCloseBody  = "Failed to close request body"
	conflictErrorMessage   = "Conflict error occurred"
	defaultErrorStatusText = "Internal Server Error"
	conflictStatusText     = "Conflict"
)

// ==================== Обработка тела запроса ====================

// readRequestBody читает и возвращает тело HTTP запроса
// Автоматически закрывает тело после чтения
func (h *UserHandler) readRequestBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(errorMessageReadBody,
			zap.String(logFieldComponent, componentRequestBody),
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		return nil, err
	}

	// Гарантируем закрытие тела запроса
	defer h.closeBody(r.Body, r.Method, r.URL.Path)

	return body, nil
}

// closeBody безопасно закрывает тело запроса с логированием ошибок
func (h *UserHandler) closeBody(body io.ReadCloser, method, path string) {
	if err := body.Close(); err != nil {
		h.logger.Error(errorMessageCloseBody,
			zap.String(logFieldComponent, componentRequestBody),
			zap.Error(err),
			zap.String("method", method),
			zap.String("path", path),
		)
	}
}

// ==================== Обработка ошибок ====================

// handleConflictError обрабатывает конфликтные ошибки (HTTP 409)
// Возвращает true, если ошибка была обработана как конфликт
func (h *UserHandler) handleConflictError(w http.ResponseWriter, r *http.Request, err error) bool {
	// TODO: Раскомментировать и доработать при реализации storage
	/*
		if h.storage.IsDuplicateError(err) {
			h.logger.Warn(conflictErrorMessage,
				zap.String("component", "conflict_handler"),
				zap.Error(err),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			// Отправляем клиенту соответствующий HTTP статус
			http.Error(w, conflictStatusText, http.StatusConflict)
			return true
		}
	*/

	// Временная заглушка для компиляции
	// После реализации storage удалить этот блок
	if false {
		h.logger.Debug("Conflict handler placeholder",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
	}

	return false
}

// writeErrorResponse записывает стандартизированный HTTP ответ с ошибкой
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	h.logger.Error("Writing error response",
		zap.Int("status_code", status),
		zap.String("error_message", message),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	http.Error(w, message, status)
}

// writeJSONResponse записывает успешный JSON ответ
func (h *UserHandler) writeJSONResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	// TODO: Реализовать сериализацию JSON и установку заголовков
	h.logger.Debug("Writing JSON response",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("data", data),
	)
}

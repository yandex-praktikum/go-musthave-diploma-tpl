package orderclient

import "fmt"

// ErrorType представляет тип ошибки API
type ErrorType string

const (
	// ErrorTypeNotFound - заказ не найден (204)
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeRateLimit - превышен лимит запросов (429)
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// ErrorTypeClient - клиентская ошибка (4xx)
	ErrorTypeClient ErrorType = "client_error"
	// ErrorTypeServer - серверная ошибка (5xx)
	ErrorTypeServer ErrorType = "server_error"
	// ErrorTypeNetwork - сетевая ошибка
	ErrorTypeNetwork ErrorType = "network_error"
	// ErrorTypeParse - ошибка парсинга ответа
	ErrorTypeParse ErrorType = "parse_error"
)

// APIError представляет ошибку API
type APIError struct {
	Type        ErrorType
	StatusCode  int
	Message     string
	RetryAfter  string
	OriginalErr error
}

func (e *APIError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s (status: %d): %v", e.Type, e.StatusCode, e.OriginalErr)
	}
	return fmt.Sprintf("%s (status: %d): %s", e.Type, e.StatusCode, e.Message)
}

// Unwrap возвращает оригинальную ошибку
func (e *APIError) Unwrap() error {
	return e.OriginalErr
}

// IsNotFound проверяет, является ли ошибка ошибкой "не найдено"
func (e *APIError) IsNotFound() bool {
	return e.Type == ErrorTypeNotFound
}

// IsRateLimit проверяет, является ли ошибка ошибкой лимита запросов
func (e *APIError) IsRateLimit() bool {
	return e.Type == ErrorTypeRateLimit
}

// Retryable проверяет, можно ли повторить запрос
func (e *APIError) Retryable() bool {
	return e.Type == ErrorTypeRateLimit ||
		e.Type == ErrorTypeServer ||
		e.Type == ErrorTypeNetwork
}

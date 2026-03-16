package orderclient

import "fmt"

const (
	// ErrorTypeNotFound - заказ не найден (204)
	ErrorTypeNotFound = "not_found"
	// ErrorTypeRateLimit - превышен лимит запросов (429)
	ErrorTypeRateLimit = "rate_limit"
	// ErrorTypeClient - клиентская ошибка (4xx)
	ErrorTypeClient = "client_error"
	// ErrorTypeServer - серверная ошибка (5xx)
	ErrorTypeServer = "server_error"
	// ErrorTypeNetwork - сетевая ошибка
	ErrorTypeNetwork = "network_error"
	// ErrorTypeParse - ошибка парсинга ответа
	ErrorTypeParse            = "parse_error"
	ErrorTypeResponseTooLarge = "RESPONSE_TOO_LARGE"
	ErrorTypeTimeout          = "TIMEOUT"
)

// APIError структура ошибки API
type APIError struct {
	Type        string
	StatusCode  int
	Message     string
	OriginalErr error
	RetryAfter  string
}

// Error возвращает строковое представление ошибки
func (e *APIError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s: %s (original: %v)", e.Type, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap возвращает оригинальную ошибку
func (e *APIError) Unwrap() error {
	return e.OriginalErr
}

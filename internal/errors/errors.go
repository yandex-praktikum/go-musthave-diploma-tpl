// Package errors предоставляет унифицированные ошибки приложения.
package errors

import "errors"

// Основные ошибки приложения
var (
	// Auth errors
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrUnauthorized       = errors.New("unauthorized")

	// Order errors
	ErrOrderAlreadyExists      = errors.New("order already exists")
	ErrInvalidOrderNumber      = errors.New("invalid order number")
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderBelongsToOtherUser = errors.New("order belongs to other user")

	// Balance errors
	ErrInsufficientFunds = errors.New("insufficient funds")

	// Accrual errors
	ErrAccrualNotFound    = errors.New("order not found in accrual system")
	ErrAccrualRateLimit   = errors.New("accrual system rate limit exceeded")
	ErrAccrualServerError = errors.New("accrual system server error")

	// Validation errors
	ErrInvalidRequest = errors.New("invalid request")
	ErrEmptyField     = errors.New("field cannot be empty")
)

// AppError унифицированная ошибка приложения
type AppError struct {
	Code    int    // HTTP статус код
	Message string // Сообщение для клиента
	Err     error  // Внутренняя ошибка
	Op      string // Операция где произошла ошибка
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Op + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError создаёт новую ошибку приложения
func NewAppError(code int, message string, op string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}

// IsAppError проверяет является ли ошибка AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetErrorCode получает HTTP код ошибки
func GetErrorCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return 500 // Internal Server Error по умолчанию
}

// GetErrorMessage получает сообщение ошибки
func GetErrorMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}
	return "Internal Server Error"
}

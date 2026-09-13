package domain

import (
	"errors"
	"fmt"
	"time"
)

// Бизнес-ошибки, возвращаемые слоями сервиса и репозиториев.
var (
	ErrLoginAlreadyExists                = errors.New("login already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrInvalidOrderNumber                = errors.New("invalid order number")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by this user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrOrderNotFound                     = errors.New("order not found")
	ErrInsufficientBalance               = errors.New("insufficient balance")
	ErrUserNotFound                      = errors.New("user not found")
	ErrUnauthorized                      = errors.New("unauthorized")
	ErrAccrualServiceUnavailable         = errors.New("accrual service unavailable")
)

// RateLimitError описывает ошибку превышения лимита запросов к accrual-сервису.
// Содержит время, через которое можно повторить запрос (из заголовка Retry-After).
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error реализует интерфейс error.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("accrual rate limit exceeded, retry after %s", e.RetryAfter)
}

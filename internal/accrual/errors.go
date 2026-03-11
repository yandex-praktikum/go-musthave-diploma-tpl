package accrual

import (
	"errors"
	"fmt"
)

// ErrOrderNotRegistered — заказ не зарегистрирован в системе расчёта (204).
type ErrOrderNotRegistered struct {
	OrderNumber string
	URL         string
}

func (e *ErrOrderNotRegistered) Error() string {
	return fmt.Sprintf("accrual: order %q not registered in accrual system (204)", e.OrderNumber)
}

// ErrRateLimit — превышен лимит запросов (429). RetryAfter — пауза в секундах перед повтором.
type ErrRateLimit struct {
	OrderNumber string
	URL         string
	RetryAfter  int // секунды из заголовка Retry-After
}

func (e *ErrRateLimit) Error() string {
	return fmt.Sprintf("accrual: rate limit exceeded for order %q", e.OrderNumber)
}

// ErrServerError — внутренняя ошибка сервера (500) или сетевая ошибка.
var ErrServerError = errors.New("accrual: server error")

// ErrServerErrorContext — обёртка над ErrServerError с контекстом запроса (для логов).
type ErrServerErrorContext struct {
	OrderNumber string
	URL         string
	Err         error
}

func (e *ErrServerErrorContext) Error() string {
	return fmt.Sprintf("accrual: order %s: %v", e.OrderNumber, e.Err)
}

func (e *ErrServerErrorContext) Unwrap() error {
	return e.Err
}

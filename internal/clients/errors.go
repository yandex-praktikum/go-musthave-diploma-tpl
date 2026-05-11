package clients

import (
	"errors"
	"fmt"
	"time"
)

var (
	// Sentinel-ошибка для быстрой проверки через errors.Is для ServiceIsBusyError
	ErrServiceIsBusy = errors.New("service responds about too many requests")
	ErrOrderNotFound = errors.New("order not found in accrual system")
)

type (
	ServiceIsBusyError struct {
		delay time.Duration
		cause error
	}

	AccrualError interface {
		Error() string
		GetDelay() time.Duration
		Is(target error) bool
		Unwrap() error
	}
)

func NewErrServiceIsBusy(delay time.Duration, cause error) *ServiceIsBusyError {
	return &ServiceIsBusyError{
		delay: delay,
		cause: cause,
	}
}

// Error делает структуру реализацией интерфейса error
func (e *ServiceIsBusyError) Error() string {
	msg := fmt.Sprintf("%s, retry after %s", ErrServiceIsBusy.Error(), e.delay)
	if e.cause != nil {
		return msg + ": " + e.cause.Error()
	}
	return msg
}

// Возвращает задержку
func (e *ServiceIsBusyError) GetDelay() time.Duration {
	return e.delay
}

// Позволяет errors.Is(err, ErrServiceIsBusy) возвращать true
func (e *ServiceIsBusyError) Is(target error) bool {
	return target == ErrServiceIsBusy
}

// Поддержка errors.Unwrap / errors.As / fmt.Errorf("%w")
// Позволяет ошибке быть частью цепочки.
func (e *ServiceIsBusyError) Unwrap() error {
	return e.cause
}

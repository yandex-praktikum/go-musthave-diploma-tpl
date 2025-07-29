package services

import (
	"errors"
	"time"
)

// Ошибки accrual сервиса
var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInternalServer    = errors.New("internal server error from accrual system")
)

// RateLimitError содержит информацию о cooldown
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimitExceeded
}

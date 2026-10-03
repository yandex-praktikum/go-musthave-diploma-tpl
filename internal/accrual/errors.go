package accrual

import "errors"

var (
	// ErrOrderNotRegistered is returned when the accrual system does not know the order.
	ErrOrderNotRegistered = errors.New("order is not registered")

	// ErrTooManyRequests is returned when the accrual system responds with HTTP 429.
	ErrTooManyRequests = errors.New("too many requests")

	// ErrAccrualServer is returned when the accrual system responds with HTTP 5xx.
	ErrAccrualServer = errors.New("the accrual server failed to fulfill an apparently valid request")

	// ErrAccrualClient is returned when the accrual system responds with HTTP 4xx.
	ErrAccrualClient = errors.New("the accrual server rejected the request")

	// ErrAccrualUnexpectedResponse is returned when the accrual system responds with an unexpected status code.
	ErrAccrualUnexpectedResponse = errors.New("unexpected response from the accrual server")
)

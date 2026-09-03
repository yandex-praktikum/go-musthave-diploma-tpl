// Package repository stores and loads Gophermart data from PostgreSQL.
package repository

import "errors"

// Repository errors used by service and handlers.
var (
	ErrUserExists             = errors.New("user already exists")
	ErrOrderExistsSameUser    = errors.New("order exists for same user")
	ErrOrderExistsAnotherUser = errors.New("order exists for another user")
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrNotFound               = errors.New("not found")
)

package usecase

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrOrderAlreadyExists  = errors.New("order already exists")
	ErrInvalidOrderNumber  = errors.New("invalid order number")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidOrderStatus  = errors.New("invalid order status")
	ErrOrderNotFound       = errors.New("order not found")
)

package repository

import "errors"

var (
	// ErrUserExists пользователь уже существует
	ErrUserExists = errors.New("user already exists")

	// ErrUserNotFound пользователь не найден
	ErrUserNotFound = errors.New("user not found")

	// ErrOrderExists заказ уже существует
	ErrOrderExists = errors.New("order already exists")

	// ErrOrderExistsByAnotherUser заказ загружен другим пользователем
	ErrOrderExistsByAnotherUser = errors.New("order exists by another user")

	// ErrOrderNotFound заказ не найден
	ErrOrderNotFound = errors.New("order not found")

	// ErrInsufficientFunds недостаточно средств
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrInternal внутренняя ошибка базы данных
	ErrInternal = errors.New("internal database error")
)

package repository

import "fmt"

// ErrDuplicateLogin — логин уже занят.
type ErrDuplicateLogin struct {
	Login string
}

func (e *ErrDuplicateLogin) Error() string {
	return fmt.Sprintf("login %q already exists", e.Login)
}

// ErrUserNotFound — пользователь не найден.
type ErrUserNotFound struct {
	Login string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with login %q not found", e.Login)
}

// ErrOrderNotFound — заказ с таким номером не найден.
type ErrOrderNotFound struct {
	Number string
}

func (e *ErrOrderNotFound) Error() string {
	return fmt.Sprintf("order with number %q not found", e.Number)
}

// ErrDuplicateWithdrawalOrder — списание по этому номеру заказа уже было (повторная подача).
type ErrDuplicateWithdrawalOrder struct {
	Order string
}

func (e *ErrDuplicateWithdrawalOrder) Error() string {
	return fmt.Sprintf("withdrawal for order %q already exists", e.Order)
}

// ErrInsufficientFunds — на счёте недостаточно средств для списания (402).
type ErrInsufficientFunds struct {
	Order string
}

func (e *ErrInsufficientFunds) Error() string {
	return fmt.Sprintf("insufficient funds for order %q", e.Order)
}

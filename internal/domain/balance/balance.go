package balance

import "errors"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
	UserId    string  `json:"-"`
}

var (
	ErrNotFound = errors.New("not found")
	//на счету недостаточно средств
	ErrInsufficientFunds = errors.New("insufficient funds")
	//неверный номер заказа
	ErrInvalidOrder   = errors.New("invalid order")
	ErrUnexpectedFlow = errors.New("user balance lost")
)

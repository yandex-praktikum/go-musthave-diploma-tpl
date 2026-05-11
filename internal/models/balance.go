package models

import "encoding/json"

type (
	Balance struct {
		Balance   Accrual `json:"current"`
		Withdrawn Accrual `json:"withdrawn"`
	}
)

func NewBalanceItem(balance, withrawn Accrual) *Balance {
	return &Balance{Balance: balance, Withdrawn: withrawn}
}

/*
Сериализует объект баланса в строку JSON
Может вернуть ошибку, коли та случится.
*/
func (m *Balance) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

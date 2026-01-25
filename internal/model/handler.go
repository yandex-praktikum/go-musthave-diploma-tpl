package model

import "encoding/json"

type UserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserOrderRes struct {
	Order   string  `json:"number"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
	Created string  `json:"uploaded_at"`
}

type BalanceRes struct {
	SberThx   float64 `json:"current"`   // подумать над типом ы
	Withdrawn float64 `json:"withdrawn"` // подумать над типомы
}

type WithdrawReq struct {
	Order   string      `json:"order"`
	Value   json.Number `json:"sum"` // подумать над типом
	Created string      `json:"processed_at"`
}

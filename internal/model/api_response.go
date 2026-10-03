package model

import (
	"time"
)

// OrderResponse is the order representation returned by GET /api/user/orders.
type OrderResponse struct {
	Number       string    `json:"number"`
	Status       Status    `json:"status"`
	Accrual      float64   `json:"accrual,omitempty"`
	UploadedTime time.Time `json:"uploaded_at"`
}

// BalanceResponse is the loyalty balance returned by GET /api/user/balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

package model

import (
	"time"
)

// AccrualResponse is the order representation returned by GET /api/user/orders.
type AccrualResponse struct {
	Number       string    `json:"number"`
	Status       Status    `json:"status"`
	Amount       float64   `json:"accrual,omitempty"`
	UploadedTime time.Time `json:"uploaded_at"`
}

// BalanceResponse is the loyalty balance returned by GET /api/user/balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

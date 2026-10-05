package model

import (
	"time"
)

// OrderResponse is the order representation returned by GET /api/user/orders.
type OrderResponse struct {
	// OrderNumber is the number of the order loaded for accrual processing.
	OrderNumber string `json:"number"`
	Status      string `json:"status"`
	// Amount is the reward accrued for the order; omitted while it is unknown.
	Amount       float64   `json:"accrual,omitempty"`
	UploadedTime time.Time `json:"uploaded_at"`
}

// BalanceResponse is the loyalty balance returned by GET /api/user/balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

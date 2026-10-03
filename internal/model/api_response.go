package model

import (
	"time"
)

// OrderResponse is the order representation returned by GET /api/user/orders.
type OrderResponse struct {
	Number       string    `json:"number"`
	Status       Status    `json:"status"`
	Accrual      float64   `json:"accrual"`
	UploadedTime time.Time `json:"uploaded_at"`
}

// BalanceResponse is the loyalty balance returned by GET /api/user/balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalRequest is the request body of POST /api/user/balance/withdraw.
type WithdrawalRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}

// WithdrawalResponse is a single withdrawal record returned by GET /api/user/withdrawals.
type WithdrawalResponse struct {
	OrderNumber   string    `json:"order"`
	Sum           float64   `json:"sum"`
	ProcessedTime time.Time `json:"processed_at"`
}

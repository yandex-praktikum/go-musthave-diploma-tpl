package model

import "time"

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

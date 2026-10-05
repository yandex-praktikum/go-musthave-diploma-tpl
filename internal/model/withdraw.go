package model

import "time"

// WithdrawalRequest is the request body of POST /api/user/balance/withdraw.
type WithdrawalRequest struct {
	// OrderNumber is the number of a NEW order; loyalty points are spent
	// from the balance to pay for it.
	OrderNumber string  `json:"order"`
	Amount      float64 `json:"sum"`
}

// WithdrawalResponse is a single withdrawal record returned by GET /api/user/withdrawals.
type WithdrawalResponse struct {
	OrderNumber   string    `json:"order"`
	Amount        float64   `json:"sum"`
	ProcessedTime time.Time `json:"processed_at"`
}

// WithdrawalRegistrationResult reports whether a withdrawal was inserted and, if not, who owns it.
type WithdrawalRegistrationResult struct {
	Inserted bool
	Owner    string
}

// ValidateWithdrawalAmount reports whether the withdrawal amount is positive.
func ValidateWithdrawalAmount(amount float64) error {
	if amount <= 0 {
		return ErrNonPositiveAmount
	}
	return nil
}

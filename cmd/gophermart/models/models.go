package models

import "time"

type User struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type Order struct {
	ID          int64     `db:"id"`
	OrderNumber string    `db:"order_number"`
	UserID      int64     `db:"user_id"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
}
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type BalanceTransaction struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	OrderID   *int64    `db:"order_id"`
	Amount    float64   `db:"amount"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

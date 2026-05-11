// Package storage предоставляет модели данных и репозитории для работы с базой данных.
package storage

import (
	"time"
)

// Статусы обработки заказов
const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

// Типы транзакций
const (
	TransactionTypeCharge   = "charge"
	TransactionTypeWithdraw = "withdraw"
)

// User представляет пользователя системы
type User struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	PasswordHash []byte    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// Order представляет заказ пользователя
type Order struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	Number     string    `db:"number"`
	Status     string    `db:"status"`
	Accrual    *float64  `db:"accrual"` // nullable
	UploadedAt time.Time `db:"uploaded_at"`
}

// Transaction представляет транзакцию по счёту пользователя
type Transaction struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Type        string    `db:"type"`
	Amount      float64   `db:"amount"`
	OrderNumber *string   `db:"order_number"` // nullable
	CreatedAt   time.Time `db:"created_at"`
}

// Withdrawal представляет списание средств
type Withdrawal struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Sum         float64   `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}

// UserBalance представляет баланс пользователя
type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// OrderResponse представляет заказ в ответе API
type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// WithdrawalResponse представляет списание в ответе API
type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

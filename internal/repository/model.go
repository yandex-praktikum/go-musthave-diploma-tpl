package repository

import (
	"time"

	"github.com/google/uuid"
)

// User модель пользователя для репозитория
type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	PassHash string    `db:"pass_hash"`
}

// Order модель заказа для репозитория
type Order struct {
	ID         int64     `db:"id"`
	Number     string    `db:"number"`
	UserID     uuid.UUID `db:"user_id"`
	Status     string    `db:"status"`
	Accrual    float64   `db:"accrual"`
	UploadedAt time.Time `db:"uploaded_at"`
}

// Balance модель баланса для репозитория
type Balance struct {
	UserID    uuid.UUID `db:"user_id"`
	Current   float64   `db:"current_balance"`
	Withdrawn float64   `db:"withdrawn"`
}

// Withdrawal модель списания для репозитория
type Withdrawal struct {
	ID          int64     `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Sum         float64   `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}

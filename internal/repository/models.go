package repository

import (
	"time"

	"github.com/true18/go-musthave-diploma-tpl/internal/money"
)

// Order processing statuses stored by Gophermart.
const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

// User is an application user stored in PostgreSQL.
type User struct {
	ID           int64
	Login        string
	PasswordHash []byte
	CreatedAt    time.Time
}

// Order is a user order stored in PostgreSQL.
type Order struct {
	Number     string
	UserID     int64
	Status     string
	Accrual    *money.Amount
	UploadedAt time.Time
}

// Balance is the calculated user loyalty balance.
type Balance struct {
	Current   money.Amount
	Withdrawn money.Amount
}

// Withdrawal is a persisted loyalty points withdrawal.
type Withdrawal struct {
	OrderNumber string
	Sum         money.Amount
	ProcessedAt time.Time
}

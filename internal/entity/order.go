package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type Order struct {
	ID         int
	UserID     int
	Status     string
	Number     string
	Accrual    decimal.Decimal
	UploadedAt time.Time
	UpdatedAt  time.Time
}

type Orders []*Order

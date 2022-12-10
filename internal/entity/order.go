package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type Order struct {
	ID          int
	UserID      int
	Status      string
	OrderNumber string
	Amount      decimal.Decimal
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

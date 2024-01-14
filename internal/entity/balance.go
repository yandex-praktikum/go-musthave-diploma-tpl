package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type Balance struct {
	ID        int
	UserID    int
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
	UpdatedAt time.Time
}

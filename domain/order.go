package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID     int64           `json:"number"      db:"id"`
	CrDt   time.Time       `json:"uploaded_at" db:"cr_dt"`
	Status string          `json:"status"      db:"status"`
	Amount decimal.Decimal `json:"accrual"     db:"amount"`
}

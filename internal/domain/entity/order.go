package entity

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderNumber string

type OrderDB struct {
	ID         int
	Number     string
	Status     string
	Accrual    sql.NullFloat64
	UploadedAt pgtype.Timestamp
	UserID     int
}

type OrderStatusJSON struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

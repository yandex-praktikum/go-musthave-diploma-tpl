package entity

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type WithdrawalDB struct {
	ID          int
	Order       string
	Sum         float64
	ProcessedAt pgtype.Timestamp
	UserID      int
}

type WithdrawalRawRecord struct {
	Order  string
	Sum    float64
	UserID int
}

type WithdrawalJSON struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

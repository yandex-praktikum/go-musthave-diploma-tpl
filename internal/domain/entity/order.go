package entity

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderNumber int

type OrderDB struct {
	ID         int
	Number     string
	Status     string
	Accrual    sql.NullInt32
	UploadedAt pgtype.Timestamp
	UserId     int
}

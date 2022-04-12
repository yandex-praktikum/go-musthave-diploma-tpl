package methods

import (
	"errors"
	db "github.com/EestiChameleon/GOphermart/internal/app/storage"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"time"
)

var (
	ErrOrderInsertFailed = errors.New("failed to save new order")
	ErrOrderUpdateFailed = errors.New("failed to update order")
)

type Order struct {
	Number     string          `json:"number"`
	UserID     int             `json:"user_id"`
	UploadedAt time.Time       `json:"uploaded_at"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
}

func NewOrder(number string) *Order {
	return &Order{
		Number:     number,
		UserID:     db.Pool.ID,
		UploadedAt: time.Now(),
		Status:     "NEW",
		Accrual:    decimal.Decimal{},
	}
}

func (o *Order) GetByNumber() error {
	err := pgxscan.Get(ctx, db.Pool.DB, o,
		"SELECT user_id, uploaded_at, status FROM orders WHERE number=$1", o.Number)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrNotFound
		}
		return err
	}

	return nil
}

func (o *Order) Add() error {
	tag, err := db.Pool.DB.Exec(ctx,
		"INSERT INTO orders(number, user_id, uploaded_at, status) "+
			"VALUES ($1, $2, $3, $4) ON CONFLICT (number) DO NOTHING;",
		o.Number, o.UserID, o.UploadedAt, o.Status)

	if err != nil {
		return err
	}

	if tag.RowsAffected() < 1 {
		return ErrOrderInsertFailed
	}

	return nil
}

func (o *Order) UpdateStatus(status string) error {
	tag, err := db.Pool.DB.Exec(ctx,
		"UPDATE orders SET status = $1 WHERE number = $2;",
		status, o.Number)

	if err != nil {
		return err
	}

	if tag.RowsAffected() < 1 {
		return ErrOrderUpdateFailed
	}

	return nil
}

func (o *Order) SetAccrual(value decimal.Decimal) error {
	tag, err := db.Pool.DB.Exec(ctx,
		"UPDATE orders SET accrual = $1 WHERE number = $2;",
		value, o.Number)

	if err != nil {
		return err
	}

	if tag.RowsAffected() < 1 {
		return ErrOrderUpdateFailed
	}

	return nil
}

func GetOrdersListByUserID() ([]*Order, error) {
	var list []*Order
	err := pgxscan.Select(ctx, db.Pool.DB, &list,
		"SELECT * FROM orders WHERE user_id=$1", db.Pool.ID)
	return list, err
}

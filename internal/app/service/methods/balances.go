package methods

import (
	"errors"
	resp "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/router/responses"
	db "github.com/EestiChameleon/BonusServiceGOphermart/internal/app/storage"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/shopspring/decimal"
	"time"
)

var (
	ErrBalanceInsertFailed = errors.New("failed to save new balance record")
)

type Balance struct {
	ID          int             `json:"id"`
	UserID      int             `json:"user_id"`
	ProcessedAt time.Time       `json:"processed_at"`
	Income      decimal.Decimal `json:"income"`
	Outcome     decimal.Decimal `json:"outcome"`
	OrderNumber string          `json:"order_number"`
}

func NewBalanceRecord() *Balance {
	return &Balance{
		ID:          0,
		UserID:      db.Pool.ID,
		ProcessedAt: time.Now(),
		Income:      decimal.Decimal{},
		Outcome:     decimal.Decimal{},
		OrderNumber: "",
	}
}

func (b *Balance) Add() error {
	err := db.Pool.DB.QueryRow(ctx,
		"INSERT INTO balances(user_id, processed_at, income, outcome, order_number) "+
			"VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		b.UserID, b.ProcessedAt, b.Income, b.Outcome, b.OrderNumber).Scan(&b.ID)

	if err != nil {
		return err
	}

	if b.ID < 1 {
		return ErrBalanceInsertFailed
	}

	return nil
}

func (b *Balance) GetBalanceAndWithdrawnByUserID() (*resp.BalanceData, error) {
	var c, w decimal.NullDecimal
	if err := db.Pool.DB.QueryRow(ctx,
		"SELECT sum(income)-sum(outcome) as current, sum(outcome) as withdraw FROM balances WHERE user_id=$1;",
		b.UserID).Scan(&c, &w); err != nil {
		return nil, err
	}

	return &resp.BalanceData{
		Current:   c.Decimal,
		Withdrawn: w.Decimal,
	}, nil
}

func GetUserWithdrawals(dest interface{}) error {
	return pgxscan.Select(ctx, db.Pool.DB, dest,
		"SELECT order_number as order, outcome as sum, processed_at "+
			"FROM balances WHERE outcome != 0 AND user_id=$1;", db.Pool.ID)
}

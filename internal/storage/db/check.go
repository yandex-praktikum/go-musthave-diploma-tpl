package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"time"
)

func (d *DateBase) CheckTableUserLogin(ctx context.Context, login string) error {
	var existingLogin string
	query := `SELECT login FROM users WHERE login = $1`

	err := d.storage.QueryRowContext(ctx, query, login).Scan(&existingLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return customerrors.ErrUserAlreadyExists
}

func (d *DateBase) CheckTableUserPassword(ctx context.Context, login string) (string, bool) {
	var existingPassword string
	query := `SELECT password FROM users WHERE login = $1`

	err := d.storage.QueryRowContext(ctx, query, login).Scan(&existingPassword)

	if err != nil {
		return "", false
	}

	if existingPassword == "" {
		return "", false
	}
	return existingPassword, true
}

func (d *DateBase) CheckWriteOffOfFunds(ctx context.Context, order string, sum float64, now time.Time) error {
	var user int
	var sumBonus float64
	queryCheckOrder := "SELECT user_id FROM loyalty WHERE order_id = $1"

	rowOrder, err := d.Get(queryCheckOrder, order)
	if err != nil {
		return customerrors.ErrNotData
	}

	if err = rowOrder.Scan(&user); err != nil {
		return err
	}

	queryCheckSumAccrual := "SELECT SUM(bonus) FROM loyalty WHERE user_id = $1"

	rowSum, err := d.Get(queryCheckSumAccrual, user)

	if err != nil {
		return customerrors.ErrNotData
	}

	if err = rowSum.Scan(&sumBonus); err != nil {
		return err
	}

	if sum > sumBonus {
		return customerrors.ErrNotEnoughBonuses
	}

	querySave := "UPDATE loyalty SET processed_at = $1 WHERE order_id = $2"

	rfc3339Time := now.Format(time.RFC3339)

	if err = d.Save(querySave, rfc3339Time, order); err != nil {
		return err
	}
	return nil
}

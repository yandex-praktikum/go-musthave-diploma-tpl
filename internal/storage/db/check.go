package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
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

func (d *DateBase) CheckWriteOffOfFunds(ctx context.Context, login, order string, sum float32, now time.Time) error {
	var sumBonus float32

	userID, err := d.GetLoginID(login)
	if err != nil {
		return err
	}

	queryCheckSumAccrual := "SELECT SUM(bonus) FROM loyalty WHERE user_id = $1"

	rowSum, err := d.Get(queryCheckSumAccrual, userID)

	if err != nil {
		return customerrors.ErrNotData
	}

	if err = rowSum.Scan(&sumBonus); err != nil {
		return customerrors.ErrNotBonus
	}

	if sum > sumBonus {
		return customerrors.ErrNotEnoughBonuses
	}

	// сохраняем новый ордер
	if err = d.SaveOrder(userID, order, string(models.NewOrder), now); err != nil {
		return err
	}

	// обновляем новыми данными
	querySave := "UPDATE loyalty SET processed_at = $1, withdraw = $2 WHERE order_id = $3"

	rfc3339Time := now.Format(time.RFC3339)

	if err = d.Save(querySave, rfc3339Time, sum, order); err != nil {
		return err
	}
	return nil
}

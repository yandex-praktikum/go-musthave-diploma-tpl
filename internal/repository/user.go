package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

func (r *Repo) RegisterUser(ctx context.Context, log, passHash string) error {
	_, err := r.conn.ExecContext(ctx, `
	INSERT INTO t_gophermart.t_users(s_login, s_pass_hash)
	VALUES ($1,$2);
	`, log, passHash)
	if err != nil {
		return fmt.Errorf(" ошибка при вставке нового пользователя: %s", err.Error())
	}
	return nil
}

func (r *Repo) GetInfoMyBalance(ctx context.Context, login string) (decimal.Decimal, decimal.Decimal, error) {
	var cb, tw decimal.Decimal
	row := r.conn.QueryRowContext(ctx,
		`SELECT current_balance, total_withdrawn FROM t_gophermart.get_user_stats($1)`, login)
	err := row.Scan(&cb, &tw)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	return cb, tw, nil

}

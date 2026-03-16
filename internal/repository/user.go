package repository

import (
	"context"
	"fmt"
	"musthave/internal/model"

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
func (r *Repo) GetUserList(ctx context.Context) ([]*model.User, error) {
	rows, err := r.conn.QueryContext(ctx, `SELECT s_login, s_pass_hash FROM t_gophermart.t_users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*model.User{}
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(&user.Login, &user.PassHash); err != nil {
			return nil, err
		}
		list = append(list, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil

}

package repository

import (
	"context"
	"fmt"
	"musthave/internal/model"

	"github.com/shopspring/decimal"
)

func (r *Repo) WithdrawnBalance(ctx context.Context, login, order string, amount decimal.Decimal) error {
	_, err := r.conn.ExecContext(ctx, `
	INSERT INTO t_gophermart.t_transactions(n_order, s_user, s_type, n_value)
	VALUES ($1,$2,$3,$4);
	`, order, login, "minus", amount)
	if err != nil {
		return fmt.Errorf(" ошибка при вставке новой зааписи о списании бонусов: %s", err.Error())
	}
	return nil
}
func (r *Repo) GetInfoWithdrawnBalance(ctx context.Context, login string) ([]*model.WithdrawReq, error) {
	rows, err := r.conn.QueryContext(ctx, `SELECT n_order, n_value, dt_created_at FROM t_gophermart.t_transactions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*model.WithdrawReq{}
	for rows.Next() {
		line := &model.WithdrawReq{}
		if err := rows.Scan(&line.Order, &line.Value, &line.Created); err != nil {
			return nil, err
		}
		list = append(list, line)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

package repository

import (
	"context"
	"fmt"
	"musthave/internal/model"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (r *Repo) SetTransaction(ctx context.Context, login, order, operationType string, amount decimal.Decimal) error {
	_, err := r.conn.ExecContext(ctx, `
	INSERT INTO t_gophermart.t_transactions(n_order, s_user, s_type, n_value)
	VALUES ($1,$2,$3,$4);
	`, order, login, operationType, amount)
	if err != nil {
		return fmt.Errorf(" ошибка при вставке новой транзакции: %s", err.Error())
	}
	return nil
}
func (r *Repo) GetInfoWithdrawnBalance(ctx context.Context, login string) ([]*model.WithdrawReq, error) {
	rows, err := r.conn.QueryContext(ctx, `
	    SELECT n_order, n_value, dt_created_at 
		FROM t_gophermart.t_transactions
		WHERE s_user = $1
		AND s_type = 'minus'
		ORDER BY dt_created_at DESC
		`, login)
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

func (r *Repo) CreateOrder(ctx context.Context, order int, login string) (time.Time, error) {
	var createdTime time.Time
	err := r.conn.QueryRowContext(ctx,
		`INSERT INTO t_gophermart.t_orders (n_order, s_user, s_status, s_sber_thx)
		 VALUES ($1, $2, $3, $4)
		 RETURNING dt_created_at`,
		order, login, model.NEW, model.CALCULATION,
	).Scan(&createdTime)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return time.Time{}, fmt.Errorf("ErrDuplicateOrder")
		}
		return time.Time{}, err
	}
	return createdTime, nil
}

func (r *Repo) SetStatus(ctx context.Context, order int, status string) error {
	query := `
	    UPDATE t_gophermart.t_orders 
	    SET s_status =  $1 
	    WHERE n_order = $2
	`
	_, err := r.conn.ExecContext(ctx, query, status, order)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) SetBonus(ctx context.Context, orderID int, status, accrual string) error {
	query := `
        UPDATE t_gophermart.t_orders 
        SET s_status = $1, s_sber_thx = $2
        WHERE n_order = $3
    `
	_, err := r.conn.ExecContext(ctx, query, status, accrual, orderID)
	return err
}

func (r *Repo) OrderExists(ctx context.Context, orderID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM t_gophermart.t_orders WHERE n_order = $1)`
	err := r.conn.QueryRowContext(ctx, query, orderID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования заказа: %w", err)
	}
	return exists, nil
}

func (r *Repo) GetOrderList(ctx context.Context, user string) ([]*model.Order, error) {
	rows, err := r.conn.QueryContext(ctx, `
	    SELECT n_order, s_status, s_sber_thx, dt_created_at 
		FROM t_gophermart.t_orders
		WHERE s_user = $1
		ORDER BY dt_created_at DESC
		`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orderList := []*model.Order{}
	for rows.Next() {
		order := &model.Order{}
		if err := rows.Scan(&order.OrderID, &order.Status, &order.Accural, &order.Created); err != nil {
			return nil, err
		}
		orderList = append(orderList, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orderList, nil
}

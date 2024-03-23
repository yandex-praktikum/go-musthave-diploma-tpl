package postgres

import (
	"context"
	"database/sql"
	"github.com/A-Kuklin/gophermart/internal/domain/entities"

	"github.com/google/uuid"
)

const (
	getUserIDByOrderSQL = `
		SELECT user_id 
		FROM gophermart.orders
		WHERE id = $1
`
	createOrderSQL = `
		INSERT INTO gophermart.orders (id, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, status, accrual, created_at, updated_at
`
	getOrdersSQL = `
		SELECT id, status, accrual, created_at
		FROM gophermart.orders
		WHERE user_id = $1
`
	getAccrualsSQL = `
		SELECT SUM(accrual)
		FROM gophermart.orders
		WHERE user_id = $1
`
)

type OrderPSQL struct {
	db *sql.DB
}

func NewOrderPSQL(db *sql.DB) *OrderPSQL {
	return &OrderPSQL{
		db: db,
	}
}

func (s *OrderPSQL) GetUserIDByOrder(ctx context.Context, orderID uint64) (uuid.UUID, error) {
	stmt, err := s.db.PrepareContext(ctx, getUserIDByOrderSQL)
	if err != nil {
		return uuid.Nil, err
	}
	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, orderID)

	var usrID uuid.UUID
	err = result.Scan(&usrID)
	if err != nil {
		return uuid.Nil, err
	}

	return usrID, nil
}

func (s *OrderPSQL) CreateOrder(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	stmt, err := s.db.PrepareContext(ctx, createOrderSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var orderDB entities.Order
	result := stmt.QueryRowContext(ctx, order.ID, order.UserID, order.Status)
	err = result.Scan(&orderDB.ID, &orderDB.UserID, &orderDB.Status, &orderDB.Accrual, &orderDB.CreatedAt, &orderDB.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &orderDB, nil
}

func (s *OrderPSQL) GetOrders(ctx context.Context, userID uuid.UUID) ([]entities.Order, error) {
	stmt, err := s.db.PrepareContext(ctx, getOrdersSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var orders []entities.Order
	rows, err := stmt.QueryContext(ctx, userID)
	for rows.Next() {
		var order entities.Order
		err = rows.Scan(&order.ID, &order.Status, &order.Accrual, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderPSQL) GetAccruals(ctx context.Context, userID uuid.UUID) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, getAccrualsSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var accrualCents sql.NullInt64
	result := stmt.QueryRowContext(ctx, userID)
	err = result.Scan(&accrualCents)
	if err != nil {
		return 0, err
	}

	if !accrualCents.Valid {
		accrualCents.Int64 = 0
	}

	return accrualCents.Int64, nil
}

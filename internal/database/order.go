package database

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

func (db *Database) InsertOrder(ctx context.Context, order *models.Order) error {
	query := "INSERT INTO orders (id, number, accrual, status, user_id, uploaded_at) VALUES($1,$2,$3,$4,$5,$6)"
	_, err := db.Exec(ctx, query, order.ID, order.Number, order.Accrual, order.Status, order.UserID, order.UploadedAt)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) SelectOrdersForProccesing(ctx context.Context) ([]*models.Order, error) {
	query := "SELECT id, number, accrual, status, user_id, uploaded_at FROM orders WHERE status IN ($1, $2)"
	rows, err := db.Query(ctx, query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}
	orders := []*models.Order{}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		var order models.Order
		err = rows.Scan(&order.ID, &order.Number, &order.Accrual, &order.Status, &order.UserID, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (db *Database) SelectUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	query := "SELECT id, number, accrual, status, user_id, uploaded_at FROM orders WHERE user_id = $1"
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	orders := []*models.Order{}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		var order models.Order
		err = rows.Scan(&order.ID, &order.Number, &order.Accrual, &order.Status, &order.UserID, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (db *Database) UpdateOrder(ctx context.Context, order *models.Order) error {
	if order.ID == uuid.Nil {
		return fmt.Errorf("when updating order, his id can't be empty")
	}
	query := "UPDATE orders SET number=$2, accrual=$3, status=$4, user_id=$5, uploaded_at=$6 WHERE id=$1"
	_, err := db.Exec(
		ctx,
		query,
		order.ID,
		order.Number,
		order.Accrual,
		order.Status,
		order.UserID,
		order.UploadedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := "SELECT id, number, accrual, status, user_id, uploaded_at FROM orders WHERE number = $1"
	row := db.QueryRow(ctx, query, number)
	order := models.Order{}
	err := row.Scan(&order.ID, &order.Number, &order.Accrual, &order.Status, &order.UserID, &order.UploadedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

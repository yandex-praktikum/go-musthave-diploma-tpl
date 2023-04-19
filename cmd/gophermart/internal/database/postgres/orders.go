package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
	"github.com/jackc/pgx/v4"
)

type order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Number     string    `json:"number"`
	UploadedAt time.Time `json:"uploaded_at"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
}

func (o order) toEntity() entity.Order {
	return entity.Order{
		ID:         o.ID,
		UserID:     o.UserID,
		Number:     o.Number,
		UploadedAt: o.UploadedAt,
		Status:     o.Status,
		Accrual:    o.Accrual,
	}
}

func (r *Repository) SaveOrder(ctx context.Context, order entity.Order) error {
	sqlCreateOrder := `INSERT INTO orders (id,user_id, number, status, accrual) VALUES ($1, $2, $3, $4,$5)`
	_, err := r.db.Exec(ctx, sqlCreateOrder, order.ID, order.UserID, order.Number, order.Status, order.Accrual)
	return err
}

func (r *Repository) GetOrder(ctx context.Context, orderNum string) (entity.Order, error) {
	var order order
	queryGetOrder := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`
	result := r.db.QueryRow(ctx, queryGetOrder, orderNum)
	if err := result.Scan(&order.ID, &order.Number, &order.UploadedAt, &order.Status, &order.Accrual); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Order{}, apperrors.ErrOrderExists
		}

		return entity.Order{}, err
	}

	return order.toEntity(), nil

}

func (r *Repository) GetUserOrders(ctx context.Context, userUID string) ([]entity.Order, error) {
	var result []entity.Order
	queryGetOrders := `SELECT number, status, uploaded_at, accrual FROM orders
					 WHERE user_id = $1 ORDER BY uploaded_at`
	rows, err := r.db.Query(ctx, queryGetOrders, userUID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var order order
		err = rows.Scan(&order.ID, &order.Number, &order.UploadedAt, &order.Status, &order.Accrual)
		if err != nil {
			return nil, err
		}
		result = append(result, order.toEntity())
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, err
}

func (r *Repository) GetAllOrders(ctx context.Context) ([]entity.Order, error) {
	var result []entity.Order
	var query string
	query = "'select user_id,  number, uploaded_at, status, accrual  from orders where status='NEW' OR status='PROCESSING'"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var res order
		err = rows.Scan(&res.ID, &res.UserID, &res.Number, &res.UploadedAt, res.Status, &res.Accrual)
		if err != nil {
			return nil, err
		}
		result = append(result, res.toEntity())
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) UpdateOrders(ctx context.Context, orders []entity.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	query := `UPDATE orders SET accrual = $1, status = $2 WHERE number = $3`
	defer tx.Rollback(ctx)
	for _, value := range orders {
		_, err = tx.Exec(ctx, query, value.Accrual, value.Status, value.Number)
		if err != nil {
			return err
		}
		tx.Commit(ctx)
	}
	return nil
}

package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ошибки репозитория заказов
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
)

// OrderRepository предоставляет методы для работы с заказами
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository создаёт новый репозиторий заказов
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create создаёт новый заказ
// Возвращает ErrOrderAlreadyExists если номер заказа уже занят
func (r *OrderRepository) Create(ctx context.Context, userID int64, number string) (*Order, error) {
	query := `
		INSERT INTO orders (user_id, number, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, number, status, accrual, uploaded_at
	`

	var order Order
	err := r.pool.QueryRow(ctx, query, userID, number, OrderStatusNew).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		// Проверяем на нарушение уникальности номера заказа
		if isOrderUniqueViolation(err) {
			return nil, ErrOrderAlreadyExists
		}
		return nil, fmt.Errorf("create order: %w", err)
	}

	return &order, nil
}

// GetByNumber находит заказ по номеру
func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (*Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order Order
	err := r.pool.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by number: %w", err)
	}

	return &order, nil
}

// GetByUser находит все заказы пользователя, отсортированные по времени загрузки (от новых к старым)
func (r *OrderRepository) GetByUser(ctx context.Context, userID int64) ([]Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders by user: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

// GetByUserAndNumber находит заказ пользователя по номеру
func (r *OrderRepository) GetByUserAndNumber(ctx context.Context, userID int64, number string) (*Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1 AND number = $2
	`

	var order Order
	err := r.pool.QueryRow(ctx, query, userID, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by user and number: %w", err)
	}

	return &order, nil
}

// UpdateStatus обновляет статус заказа и сумму начисления
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, status string, accrual *float64) error {
	query := `
		UPDATE orders
		SET status = $2, accrual = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, status, accrual)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// GetOrdersByStatus находит заказы по статусу
func (r *OrderRepository) GetOrdersByStatus(ctx context.Context, status string, limit int) ([]Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE status = $1
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("get orders by status: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

// isOrderUniqueViolation проверяет нарушение уникальности номера заказа
func isOrderUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Проверяем нарушения уникальности номера заказа (код 23505 - unique_violation)
		if pgErr.Code == "23505" {
			return pgErr.ConstraintName == "idx_orders_number" ||
				pgErr.ConstraintName == "idx_orders_user_number" ||
				pgErr.ConstraintName == "orders_number_key"
		}
	}
	return false
}

// OrderProcessingJob представляет задание на обработку заказов
type OrderProcessingJob struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    *int64
	UploadedAt time.Time
}

// GetNewOrdersForProcessing получает новые заказы для обработки
func (r *OrderRepository) GetNewOrdersForProcessing(ctx context.Context, limit int) ([]OrderProcessingJob, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE status = $1
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, OrderStatusNew, limit)
	if err != nil {
		return nil, fmt.Errorf("get new orders: %w", err)
	}
	defer rows.Close()

	var jobs []OrderProcessingJob
	for rows.Next() {
		var job OrderProcessingJob
		if err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.Number,
			&job.Status,
			&job.Accrual,
			&job.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order jobs: %w", err)
	}

	return jobs, nil
}

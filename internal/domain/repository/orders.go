package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
)

func (r *orderRepo) Create(ctx context.Context, userID int, number string, status entity.OrderStatus) (*entity.Order, error) {
	query := `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, uploaded_at, processed_at, number, status, accrual, user_id
	`

	order := &entity.Order{}
	err := r.scanOrder(
		r.db.QueryRow(ctx, query, number, userID, status),
		order,
	)
	if err != nil {
		if IsDuplicateError(err) {
			return nil, fmt.Errorf("%w: order with number '%s' already exists", ErrDuplicateKey, number)
		}
		if IsForeignKeyError(err) {
			return nil, fmt.Errorf("%w: user with ID %d not found", ErrForeignKey, userID)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// GetByID получает заказ по ID
func (r *orderRepo) GetByID(ctx context.Context, id int) (*entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE id = $1
	`

	order := &entity.Order{}
	err := r.scanOrder(
		r.db.QueryRow(ctx, query, id),
		order,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: order with ID %d", ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return order, nil
}

// GetByNumber получает заказ по номеру
func (r *orderRepo) GetByNumber(ctx context.Context, number string) (*entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE number = $1
	`

	order := &entity.Order{}
	err := r.scanOrder(
		r.db.QueryRow(ctx, query, number),
		order,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: order with number '%s'", ErrNotFound, number)
		}
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}

	return order, nil
}

// GetByUserID получает заказы пользователя
func (r *orderRepo) GetByUserID(ctx context.Context, userID int) ([]entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// GetAll получает все заказы с пагинацией
func (r *orderRepo) GetAll(ctx context.Context, limit, offset int) ([]entity.OrderWithUser, error) {
	query := `
		SELECT 
			o.id, o.uploaded_at, o.processed_at, o.number, o.status, o.accrual, o.user_id,
			u.id, u.created_at, u.updated_at, u.login, u.password, u.is_active
		FROM orders o
		JOIN users u ON o.user_id = u.id
		ORDER BY o.uploaded_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.OrderWithUser
	for rows.Next() {
		var order entity.OrderWithUser
		if err := r.scanOrderWithUser(rows, &order); err != nil {
			return nil, fmt.Errorf("failed to scan order with user: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// Update обновляет заказ
func (r *orderRepo) Update(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET 
			number = $1,
			status = $2,
			accrual = $3,
			user_id = $4,
			processed_at = NOW()
		WHERE id = $5
		RETURNING processed_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		order.Number,
		order.Status,
		order.Accrual,
		order.UserID,
		order.ID,
	).Scan(&updatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: order with ID %d", ErrNotFound, order.ID)
		}
		if IsForeignKeyError(err) {
			return fmt.Errorf("%w: user with ID %d not found", ErrForeignKey, order.UserID)
		}
		return fmt.Errorf("failed to update order: %w", err)
	}

	order.ProcessedAt = updatedAt.Format(time.RFC3339)
	return nil
}

// UpdateStatus обновляет статус заказа
func (r *orderRepo) UpdateStatus(ctx context.Context, orderID int, status entity.OrderStatus) error {
	query := `
		UPDATE orders
		SET 
			status = $1,
			processed_at = NOW()
		WHERE id = $2
		RETURNING processed_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query, status, orderID).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: order with ID %d", ErrNotFound, orderID)
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// UpdateAccrual обновляет начисление и статус заказа
func (r *orderRepo) UpdateAccrual(ctx context.Context, orderID int, accrual float64, status entity.OrderStatus) error {
	query := `
		UPDATE orders
		SET 
			accrual = $1,
			status = $2,
			processed_at = NOW()
		WHERE id = $3
		RETURNING processed_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query, accrual, status, orderID).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: order with ID %d", ErrNotFound, orderID)
		}
		return fmt.Errorf("failed to update order accrual: %w", err)
	}

	return nil
}

// Delete удаляет заказ по ID
func (r *orderRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: order with ID %d", ErrNotFound, id)
	}

	return nil
}

// DeleteByNumber удаляет заказ по номеру
func (r *orderRepo) DeleteByNumber(ctx context.Context, number string) error {
	query := `DELETE FROM orders WHERE number = $1`

	result, err := r.db.Exec(ctx, query, number)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: order with number '%s'", ErrNotFound, number)
	}

	return nil
}

// GetPending получает ожидающие обработки заказы
func (r *orderRepo) GetPending(ctx context.Context, limit int) ([]entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY uploaded_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// GetByStatus получает заказы по статусу
func (r *orderRepo) GetByStatus(ctx context.Context, status entity.OrderStatus) ([]entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE status = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders by status: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// Exists проверяет существование заказа по номеру
func (r *orderRepo) Exists(ctx context.Context, number string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, number).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}

	return exists, nil
}

// GetForProcessing получает заказы для обработки
func (r *orderRepo) GetForProcessing(ctx context.Context, limit int) ([]entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY uploaded_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query,
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders for processing: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// GetActive получает активные заказы
func (r *orderRepo) GetActive(ctx context.Context) ([]entity.Order, error) {
	query := `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY uploaded_at ASC
	`

	rows, err := r.db.Query(ctx, query,
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query active orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

// UpdateStatusByNumber обновляет статус заказа по номеру
func (r *orderRepo) UpdateStatusByNumber(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error {
	query := `
		UPDATE orders
		SET 
			status = $1,
			accrual = COALESCE($2, accrual),
			processed_at = NOW()
		WHERE number = $3
		RETURNING processed_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query, status, accrual, number).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: order with number '%s'", ErrNotFound, number)
		}
		return fmt.Errorf("failed to update order status by number: %w", err)
	}

	return nil
}

// GetOrderWithUser получает заказ с информацией о пользователе
func (r *orderRepo) GetOrderWithUser(ctx context.Context, orderID int) (*entity.OrderWithUser, error) {
	query := `
		SELECT 
			o.id, o.uploaded_at, o.processed_at, o.number, o.status, o.accrual, o.user_id,
			u.id, u.created_at, u.updated_at, u.login, u.password, u.is_active
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`

	var order entity.OrderWithUser
	if err := r.scanOrderWithUser(r.db.QueryRow(ctx, query, orderID), &order); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: order with ID %d", ErrNotFound, orderID)
		}
		return nil, fmt.Errorf("failed to get order with user: %w", err)
	}

	return &order, nil
}

// GetOrdersCount получает общее количество заказов
func (r *orderRepo) GetOrdersCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM orders`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get orders count: %w", err)
	}

	return count, nil
}

// GetOrdersSummary получает сводку по заказам
func (r *orderRepo) GetOrdersSummary(ctx context.Context) (*entity.OrdersSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'NEW' THEN 1 END) as new_count,
			COUNT(CASE WHEN status = 'PROCESSING' THEN 1 END) as processing_count,
			COUNT(CASE WHEN status = 'PROCESSED' THEN 1 END) as processed_count,
			COUNT(CASE WHEN status = 'INVALID' THEN 1 END) as invalid_count,
			COALESCE(SUM(CASE WHEN status = 'PROCESSED' THEN accrual ELSE 0 END), 0) as total_accrual
		FROM orders
	`

	summary := &entity.OrdersSummary{}
	err := r.db.QueryRow(ctx, query).Scan(
		&summary.Total,
		&summary.NewCount,
		&summary.ProcessingCount,
		&summary.ProcessedCount,
		&summary.InvalidCount,
		&summary.TotalAccrual,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders summary: %w", err)
	}

	return summary, nil
}

// ==================== Вспомогательные методы для OrderRepository ====================

// scanOrder сканирует заказ из Row или Rows
func (r *orderRepo) scanOrder(scanner Scanner, order *entity.Order) error {
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&order.ID,
		&createdAt,
		&updatedAt,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserID,
	)
	if err != nil {
		return err
	}

	order.UploadedAt = createdAt.Format(time.RFC3339)
	order.ProcessedAt = updatedAt.Format(time.RFC3339)

	return nil
}

// scanOrders сканирует список заказов
func (r *orderRepo) scanOrders(rows pgx.Rows) ([]entity.Order, error) {
	var orders []entity.Order

	for rows.Next() {
		var order entity.Order
		if err := r.scanOrder(rows, &order); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// scanOrderWithUser сканирует заказ с информацией о пользователе
func (r *orderRepo) scanOrderWithUser(scanner Scanner, orderWithUser *entity.OrderWithUser) error {
	var orderCreatedAt, orderUpdatedAt, userCreatedAt, userUpdatedAt time.Time

	err := scanner.Scan(
		&orderWithUser.ID,
		&orderCreatedAt,
		&orderUpdatedAt,
		&orderWithUser.Number,
		&orderWithUser.Status,
		&orderWithUser.Accrual,
		&orderWithUser.UserID,
		&orderWithUser.User.ID,
		&userCreatedAt,
		&userUpdatedAt,
		&orderWithUser.User.Login,
		&orderWithUser.User.Password,
		&orderWithUser.User.IsActive,
	)
	if err != nil {
		return err
	}

	orderWithUser.UploadedAt = orderCreatedAt.Format(time.RFC3339)
	orderWithUser.ProcessedAt = orderUpdatedAt.Format(time.RFC3339)
	orderWithUser.User.CreatedAt = userCreatedAt.Format(time.RFC3339)
	orderWithUser.User.UpdatedAt = userUpdatedAt.Format(time.RFC3339)

	return nil
}

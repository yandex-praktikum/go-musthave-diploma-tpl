package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"time"
)

// CreateOrder создает новый заказ
func (r *Repository) CreateOrder(ctx context.Context, userID int, number string, status entity.OrderStatus) (*entity.Order, error) {
	query := `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at, number, status, accrual, user_id
	`
	var createdAt, updatedAt time.Time
	order := &entity.Order{}
	err := r.db.QueryRow(ctx, query, number, userID, status).Scan(
		&order.ID,
		&createdAt,
		&updatedAt,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserID,
	)
	order.CreatedAt = createdAt.Format(time.RFC3339)
	order.UpdatedAt = updatedAt.Format(time.RFC3339)

	if err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateOrderAccrual обновляет начисление заказа и статус
func (r *Repository) UpdateOrderAccrual(ctx context.Context, orderID int, accrual float64, status entity.OrderStatus) error {
	query := `
        UPDATE orders
        SET 
            accrual = $1,
            status = $2,
            updated_at = NOW()
        WHERE id = $3
    `

	result, err := r.db.Exec(ctx, query, accrual, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order accrual: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// GetOrderByID получает заказ по ID
func (r *Repository) GetOrderByID(ctx context.Context, id int) (*entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE id = $1
	`

	var createdAt, updatedAt time.Time
	order := &entity.Order{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&createdAt,
		&updatedAt,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserID,
	)
	order.CreatedAt = createdAt.Format(time.RFC3339)
	order.UpdatedAt = updatedAt.Format(time.RFC3339)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetOrderByNumber получает заказ по номеру
func (r *Repository) GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE number = $1
	`

	var createdAt, updatedAt time.Time
	order := &entity.Order{}
	err := r.db.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&createdAt,
		&updatedAt,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UserID,
	)
	order.CreatedAt = createdAt.Format(time.RFC3339)
	order.UpdatedAt = updatedAt.Format(time.RFC3339)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetOrdersByUserID получает все заказы пользователя
func (r *Repository) GetOrdersByUserID(ctx context.Context, userID int) ([]entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&order.ID,
			&createdAt,
			&updatedAt,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UserID,
		)
		order.CreatedAt = createdAt.Format(time.RFC3339)
		order.UpdatedAt = updatedAt.Format(time.RFC3339)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetAllOrders получает все заказы с информацией о пользователях
func (r *Repository) GetAllOrders(ctx context.Context, limit, offset int) ([]entity.OrderWithUser, error) {
	query := `
		SELECT 
			o.id, o.created_at, o.updated_at, o.number, o.status, o.accrual, o.user_id,
			u.id, u.created_at, u.updated_at, u.login, u.password, u.is_active
		FROM orders o
		JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.OrderWithUser
	for rows.Next() {
		var orderWithUser entity.OrderWithUser
		var orderCreatedAt, orderUpdatedAt, userCreatedAt, userUpdatedAt time.Time

		err := rows.Scan(
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

		orderWithUser.CreatedAt = orderCreatedAt.Format(time.RFC3339)
		orderWithUser.UpdatedAt = orderUpdatedAt.Format(time.RFC3339)
		orderWithUser.User.CreatedAt = userCreatedAt.Format(time.RFC3339)
		orderWithUser.User.UpdatedAt = userUpdatedAt.Format(time.RFC3339)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, orderWithUser)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetPendingOrders получает заказы в обработке
func (r *Repository) GetPendingOrders(ctx context.Context, limit int) ([]entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&order.ID,
			&createdAt,
			&updatedAt,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UserID,
		)
		order.CreatedAt = createdAt.Format(time.RFC3339)
		order.UpdatedAt = updatedAt.Format(time.RFC3339)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetOrdersByStatus получает заказы по статусу
func (r *Repository) GetOrdersByStatus(ctx context.Context, status entity.OrderStatus) ([]entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders by status: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&order.ID,
			&createdAt,
			&updatedAt,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UserID,
		)
		order.CreatedAt = createdAt.Format(time.RFC3339)
		order.UpdatedAt = updatedAt.Format(time.RFC3339)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
func (r *Repository) UpdateOrderStatus(ctx context.Context, orderID int, status entity.OrderStatus) error {
	query := `
		UPDATE orders
		SET 
			status = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// UpdateOrder обновляет заказ
func (r *Repository) UpdateOrder(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET 
			number = $1,
			status = $2,
			accrual = $3,
			user_id = $4,
			updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.Exec(ctx, query,
		order.Number,
		order.Status,
		order.Accrual,
		order.UserID,
		order.ID,
	)

	if err != nil {
		if IsForeignKeyError(err) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// GetActiveOrders возвращает активные заказы
func (r *Repository) GetActiveOrders(ctx context.Context) ([]*entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, entity.OrderStatusNew, entity.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to query active orders: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&order.ID,
			&createdAt,
			&updatedAt,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UserID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		order.CreatedAt = createdAt.Format(time.RFC3339)
		if !updatedAt.IsZero() {
			order.UpdatedAt = updatedAt.Format(time.RFC3339)
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetOrdersForProcessing возвращает заказы для обработки
func (r *Repository) GetOrdersForProcessing(ctx context.Context, limit int) ([]*entity.Order, error) {
	query := `
		SELECT id, created_at, updated_at, number, status, accrual, user_id
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY created_at ASC
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

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&order.ID,
			&createdAt,
			&updatedAt,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UserID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		order.CreatedAt = createdAt.Format(time.RFC3339)
		if !updatedAt.IsZero() {
			order.UpdatedAt = updatedAt.Format(time.RFC3339)
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// ExistsByNumber проверяет существование заказа по номеру
func (r *Repository) ExistsByNumber(ctx context.Context, number string) (bool, error) {
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

// Delete удаляет заказ
func (r *Repository) Delete(ctx context.Context, number string) error {
	query := `DELETE FROM orders WHERE number = $1`

	result, err := r.db.Exec(ctx, query, number)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", number)
	}

	return nil
}

// UpdateStatus обновляет статус заказа
func (r *Repository) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3
	`

	result, err := r.db.Exec(ctx, query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", number)
	}

	return nil
}

func (r *Repository) GetWithdrawals(ctx context.Context, userID string, status entity.OrderStatus) ([]entity.Withdraw, error) {

	query := `
		SELECT number, accrual, updated_at 
		FROM orders 
		WHERE user_id = $1 AND status = $2 AND accrual < 0 
		ORDER BY  created_at ASC
	`
	rows, err := r.db.Query(ctx, query, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders for withdrawals: %w", err)
	}
	defer rows.Close()
	var withdraws []entity.Withdraw
	for rows.Next() {
		var withdraw entity.Withdraw
		var updatedAt time.Time
		errScanRows := rows.Scan(&withdraw.Order, &withdraw.Sum, &updatedAt)
		if errScanRows != nil {
			return nil, fmt.Errorf("failed to scan order: %w", errScanRows)
		}
		withdraw.ProcessedAt = updatedAt.Format(time.RFC3339)

		withdraws = append(withdraws, withdraw)
	}

	return withdraws, nil
}

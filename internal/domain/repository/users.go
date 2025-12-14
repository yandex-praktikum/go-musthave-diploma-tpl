package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
)

// ==================== Реализация UserRepository ====================

// Create создает нового пользователя
func (r *userRepo) Create(ctx context.Context, login, password string) (*entity.User, error) {
	query := `
		INSERT INTO users (login, password, is_active)
		VALUES ($1, $2, true)
		RETURNING id, created_at, updated_at, login, password, is_active
	`

	user := &entity.User{}
	err := r.scanUser(
		r.db.QueryRow(ctx, query, login, password),
		user,
	)
	if err != nil {
		if IsDuplicateError(err) {
			return nil, fmt.Errorf("%w: login '%s' already exists", ErrDuplicateKey, login)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID получает пользователя по ID
func (r *userRepo) GetByID(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE id = $1
	`

	user := &entity.User{}
	err := r.scanUser(
		r.db.QueryRow(ctx, query, id),
		user,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	return user, nil
}

// GetByLogin получает пользователя по логину
func (r *userRepo) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE login = $1
	`

	user := &entity.User{}
	err := r.scanUser(
		r.db.QueryRow(ctx, query, login),
		user,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	return user, nil
}

// Update обновляет данные пользователя
func (r *userRepo) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET 
			login = $1,
			password = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		user.Login,
		user.Password,
		user.IsActive,
		user.ID,
	).Scan(&updatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if IsDuplicateError(err) {
			return fmt.Errorf("%w: login '%s' already exists", ErrDuplicateKey, user.Login)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.UpdatedAt = updatedAt.Format(time.RFC3339)
	return nil
}

// Delete удаляет пользователя
func (r *userRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		if IsForeignKeyError(err) {
			return fmt.Errorf("%w: user has active orders", ErrForeignKey)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdatePassword обновляет пароль пользователя
func (r *userRepo) UpdatePassword(ctx context.Context, userID int, newPassword string) error {
	query := `
		UPDATE users
		SET 
			password = $1,
			updated_at = NOW()
		WHERE id = $2 AND is_active = true
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query, newPassword, userID).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Deactivate деактивирует пользователя
func (r *userRepo) Deactivate(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET 
			is_active = false,
			updated_at = NOW()
		WHERE id = $1 AND is_active = true
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(ctx, query, userID).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// GetBalance получает баланс пользователя
func (r *userRepo) GetBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	query := `
		WITH processed_orders AS (
			SELECT accrual
			FROM orders
			WHERE user_id = $1 AND status = 'PROCESSED'
		)
		SELECT 
			$1::int as user_id,
			COALESCE(SUM(accrual), 0) as current,
			COALESCE(ABS(SUM(CASE WHEN accrual < 0 THEN accrual ELSE 0 END)), 0) as withdrawn
		FROM processed_orders
	`

	balance := &entity.UserBalance{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Если у пользователя нет заказов, возвращаем нулевой баланс
			balance.UserID = userID
			balance.Current = 0
			balance.Withdrawn = 0
			return balance, nil
		}
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// Exists проверяет существование пользователя
func (r *userRepo) Exists(ctx context.Context, userID int) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// GetStats получает статистику пользователя
func (r *userRepo) GetStats(ctx context.Context, userID int) (*entity.UserStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COUNT(CASE WHEN status = 'PROCESSED' THEN 1 END) as processed_orders,
			COUNT(CASE WHEN status = 'NEW' THEN 1 END) as new_orders,
			COUNT(CASE WHEN status = 'PROCESSING' THEN 1 END) as processing_orders,
			COUNT(CASE WHEN status = 'INVALID' THEN 1 END) as invalid_orders,
			COALESCE(SUM(CASE WHEN status = 'PROCESSED' THEN accrual ELSE 0 END), 0) as total_accrual,
			COALESCE(ABS(SUM(CASE WHEN status = 'PROCESSED' AND accrual < 0 THEN accrual ELSE 0 END)), 0) as total_withdrawn
		FROM orders
		WHERE user_id = $1
	`

	stats := &entity.UserStats{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.TotalOrders,
		&stats.ProcessedOrders,
		&stats.NewOrders,
		&stats.ProcessingOrders,
		&stats.InvalidOrders,
		&stats.TotalAccrual,
		&stats.TotalWithdrawn,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Если у пользователя нет заказов, возвращаем нулевую статистику
			return &entity.UserStats{
				UserID:           userID,
				TotalOrders:      0,
				ProcessedOrders:  0,
				NewOrders:        0,
				ProcessingOrders: 0,
				InvalidOrders:    0,
				TotalAccrual:     0,
				TotalWithdrawn:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	stats.UserID = userID
	return stats, nil
}

// GetCount получает общее количество активных пользователей
func (r *userRepo) GetCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_active = true`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get users count: %w", err)
	}

	return count, nil
}

// GetAll получает пользователей с пагинацией
func (r *userRepo) GetAll(ctx context.Context, limit, offset int) ([]entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE is_active = true
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := r.scanUser(rows, &user); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// GetActiveUsers получает всех активных пользователей
func (r *userRepo) GetActiveUsers(ctx context.Context) ([]entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE is_active = true
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := r.scanUser(rows, &user); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// GetUserWithOrders получает пользователя с его заказами
func (r *userRepo) GetUserWithOrders(ctx context.Context, userID int) (*entity.UserWithOrders, error) {
	// Получаем пользователя
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Получаем заказы пользователя
	orders, err := r.db.Query(ctx, `
		SELECT id, uploaded_at, processed_at, number, status, accrual, user_id
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer orders.Close()

	var orderList []entity.Order
	for orders.Next() {
		var order entity.Order
		if err := r.scanOrder(orders, &order); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orderList = append(orderList, order)
	}

	if err := orders.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &entity.UserWithOrders{
		User:   *user,
		Orders: orderList,
	}, nil
}

// ==================== Вспомогательные методы ====================

// scanUser сканирует пользователя из Row или Rows
func (r *userRepo) scanUser(scanner Scanner, user *entity.User) error {
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&user.ID,
		&createdAt,
		&updatedAt,
		&user.Login,
		&user.Password,
		&user.IsActive,
	)
	if err != nil {
		return err
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	return nil
}

// scanOrder сканирует заказ из Row или Rows
func (r *userRepo) scanOrder(scanner Scanner, order *entity.Order) error {
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

// GetUserByLogin - алиас для GetByLogin (для обратной совместимости)
func (r *userRepo) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	return r.GetByLogin(ctx, login)
}

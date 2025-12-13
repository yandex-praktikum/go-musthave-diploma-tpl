package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"time"
)

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(ctx context.Context, login, password string) (*entity.User, error) {
	query := `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at, login, password, is_active
	`
	var createdAt, updatedAt time.Time
	user := &entity.User{}
	err := r.db.QueryRow(ctx, query, login, password).Scan(
		&user.ID,
		&createdAt,
		&updatedAt,
		&user.Login,
		&user.Password,
		&user.IsActive,
	)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)

	if err != nil {
		if isDuplicateError(err) {
			return nil, fmt.Errorf("user with login '%s' already exists", login)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE id = $1 AND is_active = true
	`
	var createdAt, updatedAt time.Time
	user := &entity.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&createdAt,
		&updatedAt,
		&user.Login,
		&user.Password,
		&user.IsActive,
	)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByLogin получает пользователя по логину
func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	query := `
		SELECT id, created_at, updated_at, login, password, is_active
		FROM users
		WHERE login = $1 AND is_active = true
	`
	var createdAt, updatedAt time.Time
	user := &entity.User{}
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&createdAt,
		&updatedAt,
		&user.Login,
		&user.Password,
		&user.IsActive,
	)
	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Format(time.RFC3339)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET 
			login = $1,
			password = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query,
		user.Login,
		user.Password,
		user.IsActive,
		user.ID,
	)

	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("login '%s' already exists", user.Login)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser удаляет пользователя
func (r *Repository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		if isForeignKeyError(err) {
			return fmt.Errorf("cannot delete user with existing orders")
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateUserPassword обновляет пароль пользователя
func (r *Repository) UpdateUserPassword(ctx context.Context, userID int, newPassword string) error {
	query := `
		UPDATE users
		SET 
			password = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, newPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeactivateUser деактивирует пользователя
func (r *Repository) DeactivateUser(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET 
			is_active = false,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetUserBalance получает баланс пользователя
func (r *Repository) GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	query := `
		SELECT 
			$1::int as user_id,
			COALESCE(SUM(CASE WHEN status = 'PROCESSED' THEN accrual ELSE 0 END), 0) as current,
			COALESCE(SUM(CASE WHEN status = 'INVALID' THEN 0 ELSE accrual END), 0) as withdrawn
		FROM orders
		WHERE user_id = $1
	`

	balance := &entity.UserBalance{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

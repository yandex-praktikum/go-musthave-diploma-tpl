package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ошибки репозитория пользователей
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository предоставляет методы для работы с пользователями
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создаёт новый репозиторий пользователей
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create создаёт нового пользователя
// Возвращает ErrUserAlreadyExists если логин уже занят
func (r *UserRepository) Create(ctx context.Context, login string, passwordHash []byte) (*User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, password_hash, created_at
	`

	var user User
	err := r.pool.QueryRow(ctx, query, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		// Проверяем на нарушение уникальности логина
		if isUniqueViolation(err, "users_login_key") {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

// GetByLogin находит пользователя по логину
// Возвращает ErrUserNotFound если пользователь не найден
func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user User
	err := r.pool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return &user, nil
}

// GetByID находит пользователя по ID
// Возвращает ErrUserNotFound если пользователь не найден
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}

	return &user, nil
}

// isUniqueViolation проверяет, является ли ошибка нарушением уникальности
func isUniqueViolation(err error, constraintName string) bool {
	// Проверяем код ошибки PostgreSQL: 23505 - unique_violation
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}

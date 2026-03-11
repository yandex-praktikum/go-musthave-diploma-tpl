package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository — реализация UserRepository для PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create вставляет пользователя в БД. При нарушении UNIQUE по login возвращает *repository.ErrDuplicateLogin.
func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (*models.User, error) {
	q := `INSERT INTO users (login, password_hash) VALUES ($1, $2)
		  RETURNING id, login, password_hash, active, created_at`
	var user models.User
	err := r.db.QueryRowContext(ctx, q, login, passwordHash).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.Active, &user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &repository.ErrDuplicateLogin{Login: login}
		}
		return nil, err
	}
	return &user, nil
}

// GetByLogin возвращает пользователя по логину. Если не найден — *repository.ErrUserNotFound.
func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	q := `SELECT id, login, password_hash, active, created_at FROM users WHERE login = $1`
	var user models.User
	err := r.db.QueryRowContext(ctx, q, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.Active, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &repository.ErrUserNotFound{Login: login}
		}
		return nil, err
	}
	return &user, nil
}

package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

// UserRepository — репозиторий пользователей, работающий с таблицей users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт новый экземпляр UserRepository поверх переданного соединения с БД.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создает нового пользователя
func (r *UserRepository) Create(user *domain.User) error {
	query := `
        INSERT INTO users (login, password, created_at)
        VALUES ($1, $2, $3)
        RETURNING id
    `

	err := r.db.QueryRow(query, user.Login, user.Password, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrLoginAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByLogin находит пользователя по логину
func (r *UserRepository) FindByLogin(login string) (*domain.User, error) {
	user := &domain.User{}
	query := `
        SELECT id, login, password, created_at
        FROM users
        WHERE login = $1
    `

	err := r.db.QueryRow(query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// FindByID находит пользователя по ID
func (r *UserRepository) FindByID(id int64) (*domain.User, error) {
	user := &domain.User{}
	query := `
        SELECT id, login, password, created_at
        FROM users
        WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

// Exists проверяет, существует ли пользователь с таким логином
func (r *UserRepository) Exists(login string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)`

	err := r.db.QueryRow(query, login).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

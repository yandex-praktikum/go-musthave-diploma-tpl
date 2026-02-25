// Package postgres реализует репозиторий для работы с PostgreSQL
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/anon-d/gophermarket/internal/repository"
)

// GenericRepository базовый generic репозиторий для устранения дублирования
type GenericRepository[T any] struct {
	db *sqlx.DB
}

// NewGenericRepository создаёт новый generic репозиторий
func NewGenericRepository[T any](db *sqlx.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

// GetOne выполняет запрос и возвращает один объект
func (r *GenericRepository[T]) GetOne(ctx context.Context, query string, args ...any) (*T, error) {
	var entity T
	err := r.db.GetContext(ctx, &entity, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %w", repository.ErrInternal, err)
	}
	return &entity, nil
}

// GetMany выполняет запрос и возвращает список объектов
func (r *GenericRepository[T]) GetMany(ctx context.Context, query string, args ...any) ([]T, error) {
	var entities []T
	err := r.db.SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrInternal, err)
	}
	return entities, nil
}

// Exists проверяет существование записи
func (r *GenericRepository[T]) Exists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, args...)
	if err != nil {
		return false, fmt.Errorf("%w: %w", repository.ErrInternal, err)
	}
	return exists, nil
}

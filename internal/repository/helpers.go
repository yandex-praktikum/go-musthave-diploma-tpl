package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation проверяет, является ли ошибка нарушением уникальности
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Код 23505 - unique_violation в PostgreSQL
		return pgErr.Code == "23505"
	}
	return false
}

// isForeignKeyViolation проверяет, является ли ошибка нарушением внешнего ключа
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Код 23503 - foreign_key_violation в PostgreSQL
		return pgErr.Code == "23503"
	}
	return false
}

// extractConstraintName извлекает имя ограничения из ошибки PostgreSQL
func extractConstraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

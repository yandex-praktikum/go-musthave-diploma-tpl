package store

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorUniqueViolation struct {
	message string
}

func (e ErrorUniqueViolation) Error() string {
	return e.message
}

func handleUniqueViolation(errFromStore error) error {
	var pgErr *pgconn.PgError
	if errors.As(errFromStore, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrorUniqueViolation{}
		}
	}

	return errFromStore
}

type OrderNotExists struct {
	message string
}

func (e OrderNotExists) Error() string {
	return e.message
}

type UserNoMoney struct {
	message string
}

func (e UserNoMoney) Error() string {
	return e.message
}

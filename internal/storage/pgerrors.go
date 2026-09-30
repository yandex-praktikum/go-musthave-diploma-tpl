package storage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresErrorClassification tells whether a failed operation can be retried.
type PostgresErrorClassification int

const (
	// NonRetriable means the operation must not be retried.
	NonRetriable PostgresErrorClassification = iota

	// Retriable means the operation can be retried.
	Retriable
)

// PostgresErrorClassifier classifies PostgreSQL errors.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier returns a classifier for PostgreSQL errors.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify returns the classification of the given error.
func (c *PostgresErrorClassifier) Classify(err error) PostgresErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Try to extract pgconn.PgError from the error chain.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return ClassifyPgError(pgErr)
	}

	// Unknown errors are treated as non-retriable.
	return NonRetriable
}

// ClassifyPgError classifies a PostgreSQL error code as retriable or non-retriable.
func ClassifyPgError(pgErr *pgconn.PgError) PostgresErrorClassification {
	// PostgreSQL error codes: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Class 08 — connection exception.
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Class 40 — transaction rollback.
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Class 57 — operator intervention.
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	switch pgErr.Code {
	// Class 22 — data exception.
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Class 23 — integrity constraint violation.
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Class 42 — syntax error or access rule violation.
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	// Unknown error codes are non-retriable.
	return NonRetriable
}

package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	pgUnique := &pgconn.PgError{Code: "23505"}
	pgOther := &pgconn.PgError{Code: "23503"}
	regular := errors.New("some error")

	if !isUniqueViolation(pgUnique) {
		t.Error("isUniqueViolation(23505) = false, want true")
	}
	if isUniqueViolation(pgOther) {
		t.Error("isUniqueViolation(23503) = true, want false")
	}
	if isUniqueViolation(regular) {
		t.Error("isUniqueViolation(regular error) = true, want false")
	}
	if isUniqueViolation(nil) {
		t.Error("isUniqueViolation(nil) = true, want false")
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	pgFK := &pgconn.PgError{Code: "23503"}
	pgOther := &pgconn.PgError{Code: "23505"}
	regular := errors.New("some error")

	if !isForeignKeyViolation(pgFK) {
		t.Error("isForeignKeyViolation(23503) = false, want true")
	}
	if isForeignKeyViolation(pgOther) {
		t.Error("isForeignKeyViolation(23505) = true, want false")
	}
	if isForeignKeyViolation(regular) {
		t.Error("isForeignKeyViolation(regular error) = true, want false")
	}
	if isForeignKeyViolation(nil) {
		t.Error("isForeignKeyViolation(nil) = true, want false")
	}
}

func TestExtractConstraintName(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", ConstraintName: "users_login_key"}
	if got := extractConstraintName(pgErr); got != "users_login_key" {
		t.Errorf("extractConstraintName() = %q, want %q", got, "users_login_key")
	}
	if got := extractConstraintName(errors.New("regular")); got != "" {
		t.Errorf("extractConstraintName(regular) = %q, want empty", got)
	}
	if got := extractConstraintName(nil); got != "" {
		t.Errorf("extractConstraintName(nil) = %q, want empty", got)
	}
}

// requireError проверяет, что err оборачивает want.
func requireError(t *testing.T, err, want error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %v, got nil", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected error %v, got %v", want, err)
	}
}

// requireNotError проверяет, что err НЕ оборачивает unwanted.
func requireNotError(t *testing.T, err, unwanted error) {
	t.Helper()
	if errors.Is(err, unwanted) {
		t.Fatalf("error %v should not be %v", err, unwanted)
	}
}

// requireNoError проверяет, что err == nil.
func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// requireMockExpectations проверяет, что все ожидания sqlmock выполнены.
func requireMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

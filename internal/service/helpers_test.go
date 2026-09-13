package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
)

// newTestDB создаёт тестовое соединение с БД через sqlmock.
func newTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// newAuthFixture создаёт AuthService на sqlmock.
func newAuthFixture(t *testing.T) (*AuthService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newTestDB(t)
	return NewAuthService(repository.NewUserRepository(db)), mock
}

// newOrderFixture создаёт OrderService на sqlmock с указанным accrual-клиентом.
func newOrderFixture(t *testing.T, client *AccrualClient) (*OrderService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newTestDB(t)
	return NewOrderService(
		repository.NewOrderRepository(db),
		repository.NewBalanceRepository(db),
		client,
	), mock
}

// newBalanceFixture создаёт BalanceService на sqlmock.
func newBalanceFixture(t *testing.T) (*BalanceService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newTestDB(t)
	return NewBalanceService(
		repository.NewBalanceRepository(db),
		repository.NewWithdrawalRepository(db),
		db,
	), mock
}

// requireError проверяет, что err не nil и оборачивает ожидаемую ошибку.
// Использует t.Helper(), чтобы ошибки указывали на вызывающий тест.
func requireError(t *testing.T, err error, want error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %v, got nil", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected error %v, got %v", want, err)
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

// requireNotError проверяет, что err НЕ оборачивает нежелательную ошибку.
func requireNotError(t *testing.T, err, unwanted error) {
	t.Helper()
	if errors.Is(err, unwanted) {
		t.Fatalf("error %v should not be %v", err, unwanted)
	}
}

// requireWrapped проверяет, что err оборачивает именно want (а не nil и не другую).
func requireWrapped(t *testing.T, err, want error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected wrapped error %v, got nil", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error %v, got %v", want, err)
	}
}

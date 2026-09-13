package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

func TestBalanceService_GetBalance(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(100.0, 50.0))

	balance, err := svc.GetBalance(1)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if balance.Current != 100.0 || balance.Withdrawn != 50.0 {
		t.Errorf("unexpected balance: %+v", balance)
	}
}

func TestBalanceService_GetBalance_Error(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WillReturnError(errors.New("db down"))

	if _, err := svc.GetBalance(1); err == nil {
		t.Error("GetBalance() expected error, got nil")
	}
}

func TestBalanceService_Withdraw_InvalidOrder(t *testing.T) {
	svc, _ := newBalanceFixture(t)

	if err := svc.Withdraw(1, "bad-number", 100); !errors.Is(err, domain.ErrInvalidOrderNumber) {
		t.Errorf("Withdraw() error = %v, want ErrInvalidOrderNumber", err)
	}
}

func TestBalanceService_Withdraw_NonPositiveSum(t *testing.T) {
	svc, _ := newBalanceFixture(t)

	if err := svc.Withdraw(1, "4561261212345467", 0); err == nil {
		t.Error("Withdraw() with sum=0 expected error, got nil")
	}
	if err := svc.Withdraw(1, "4561261212345467", -10); err == nil {
		t.Error("Withdraw() with negative sum expected error, got nil")
	}
}

func TestBalanceService_Withdraw_InsufficientBalance(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	if err := svc.Withdraw(1, "4561261212345467", 100); !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Errorf("Withdraw() error = %v, want ErrInsufficientBalance", err)
	}
}

func TestBalanceService_Withdraw_Success(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow(500.0))
	mock.ExpectExec(`UPDATE balances`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`INSERT INTO withdrawals`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err := svc.Withdraw(1, "4561261212345467", 100.0)
	requireNoError(t, err)

	requireMockExpectations(t, mock)
}

func TestBalanceService_Withdraw_SubtractError(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	if err := svc.Withdraw(1, "4561261212345467", 100); err == nil {
		t.Error("Withdraw() expected error, got nil")
	}
}

func TestBalanceService_Withdraw_CreateRecordError(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow(500.0))
	mock.ExpectExec(`UPDATE balances`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`INSERT INTO withdrawals`).WillReturnError(errors.New("db down"))

	if err := svc.Withdraw(1, "4561261212345467", 100); err == nil {
		t.Error("Withdraw() expected error, got nil")
	}
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	svc, mock := newBalanceFixture(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}).
			AddRow(1, 1, "123", 100.0, now))

	withdrawals, err := svc.GetWithdrawals(1)
	if err != nil {
		t.Fatalf("GetWithdrawals() error = %v", err)
	}
	if len(withdrawals) != 1 || withdrawals[0].OrderNumber != "123" {
		t.Errorf("unexpected withdrawals: %+v", withdrawals)
	}
}

func TestBalanceService_GetWithdrawals_Error(t *testing.T) {
	svc, mock := newBalanceFixture(t)

	mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WillReturnError(errors.New("db down"))

	if _, err := svc.GetWithdrawals(1); err == nil {
		t.Error("GetWithdrawals() expected error, got nil")
	}
}

package handlers

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBalanceHandler_GetBalance_Success(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(500.5, 100.0))

	req := f.authedRequest(t, http.MethodGet, "/api/user/balance", 1)
	rec := f.serve(t, http.HandlerFunc(f.bal.GetBalance), req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"current":500.5`) {
		t.Errorf("body should contain current balance, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"withdrawn":100`) {
		t.Errorf("body should contain withdrawn, got: %s", rec.Body.String())
	}
}

func TestBalanceHandler_GetBalance_NoUserID(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()

	f.bal.GetBalance(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBalanceHandler_GetBalance_Error(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT current, withdrawn FROM balances`).
		WillReturnError(errDBDown())

	req := f.authedRequest(t, http.MethodGet, "/api/user/balance", 1)
	rec := f.serve(t, http.HandlerFunc(f.bal.GetBalance), req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestBalanceHandler_Withdraw_Success(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectBegin()
	f.mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow(500.0))
	f.mock.ExpectExec(`UPDATE balances`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectQuery(`INSERT INTO withdrawals`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	f.mock.ExpectCommit()

	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{"order":"4561261212345467","sum":100}`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestBalanceHandler_Withdraw_NoUserID(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", nil)
	rec := httptest.NewRecorder()

	f.bal.Withdraw(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBalanceHandler_Withdraw_InvalidJSON(t *testing.T) {
	f := newFixture(t)

	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{invalid`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBalanceHandler_Withdraw_InvalidFields(t *testing.T) {
	f := newFixture(t)

	// Пустой номер заказа
	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{"order":"","sum":100}`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("empty order: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	// Неположительная сумма
	req2 := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req2.Body = io.NopCloser(strings.NewReader(`{"order":"123","sum":0}`))
	rec2 := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("zero sum: status = %d, want %d", rec2.Code, http.StatusBadRequest)
	}
}

func TestBalanceHandler_Withdraw_InvalidOrderNumber(t *testing.T) {
	f := newFixture(t)

	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{"order":"123","sum":100}`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestBalanceHandler_Withdraw_InsufficientBalance(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectBegin()
	f.mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnError(sql.ErrNoRows)
	f.mock.ExpectRollback()

	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{"order":"4561261212345467","sum":100}`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)

	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusPaymentRequired)
	}
}

func TestBalanceHandler_Withdraw_InternalError(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectBegin()
	f.mock.ExpectQuery(`SELECT current FROM balances`).
		WillReturnError(errDBDown())
	f.mock.ExpectRollback()

	req := f.authedRequest(t, http.MethodPost, "/api/user/balance/withdraw", 1)
	req.Body = io.NopCloser(strings.NewReader(`{"order":"4561261212345467","sum":100}`))
	rec := f.serve(t, http.HandlerFunc(f.bal.Withdraw), req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestBalanceHandler_GetWithdrawals_Success(t *testing.T) {
	f := newFixture(t)
	now := time.Now()

	f.mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}).
			AddRow(1, 1, "4561261212345467", 100.0, now))

	req := f.authedRequest(t, http.MethodGet, "/api/user/withdrawals", 1)
	rec := f.serve(t, http.HandlerFunc(f.bal.GetWithdrawals), req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "4561261212345467") {
		t.Errorf("body should contain order number, got: %s", rec.Body.String())
	}
}

func TestBalanceHandler_GetWithdrawals_NoUserID(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rec := httptest.NewRecorder()

	f.bal.GetWithdrawals(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBalanceHandler_GetWithdrawals_Empty(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}))

	req := f.authedRequest(t, http.MethodGet, "/api/user/withdrawals", 1)
	rec := f.serve(t, http.HandlerFunc(f.bal.GetWithdrawals), req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBalanceHandler_GetWithdrawals_Error(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT id, user_id, order_number, sum, processed_at FROM withdrawals`).
		WillReturnError(errors.New("db down"))

	req := f.authedRequest(t, http.MethodGet, "/api/user/withdrawals", 1)
	rec := f.serve(t, http.HandlerFunc(f.bal.GetWithdrawals), req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

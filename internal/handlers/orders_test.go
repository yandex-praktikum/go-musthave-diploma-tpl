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
	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

// errDBDown возвращает типовую ошибку БД для тестов.
func errDBDown() error { return errors.New("db down") }

func TestOrderHandler_Upload_Success(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs("4561261212345467").
		WillReturnError(sql.ErrNoRows)
	f.mock.ExpectExec(`INSERT INTO orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = io.NopCloser(strings.NewReader("4561261212345467"))
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestOrderHandler_Upload_NoUserID(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOrderHandler_Upload_EmptyBody(t *testing.T) {
	f := newFixture(t)

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = http.NoBody
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOrderHandler_Upload_InvalidNumber(t *testing.T) {
	f := newFixture(t)

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = io.NopCloser(strings.NewReader("123"))
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestOrderHandler_Upload_AlreadyByUser(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("4561261212345467", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = io.NopCloser(strings.NewReader("4561261212345467"))
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestOrderHandler_Upload_AlreadyByAnother(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("4561261212345467", 2, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = io.NopCloser(strings.NewReader("4561261212345467"))
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestOrderHandler_Upload_InternalError(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(errDBDown())

	req := f.authedRequest(t, http.MethodPost, "/api/user/orders", 1)
	req.Body = io.NopCloser(strings.NewReader("4561261212345467"))
	rec := f.serve(t, http.HandlerFunc(f.ord.Upload), req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestOrderHandler_List_Success(t *testing.T) {
	f := newFixture(t)
	now := time.Now()
	accrual := 100.0

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("4561261212345467", 1, domain.OrderStatusProcessed, accrual, now, now))

	req := f.authedRequest(t, http.MethodGet, "/api/user/orders", 1)
	rec := f.serve(t, http.HandlerFunc(f.ord.List), req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rec.Body.String(), "4561261212345467") {
		t.Errorf("body should contain order number, got: %s", rec.Body.String())
	}
}

func TestOrderHandler_List_NoUserID(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := f.serve(t, http.HandlerFunc(f.ord.List), req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOrderHandler_List_Empty(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}))

	req := f.authedRequest(t, http.MethodGet, "/api/user/orders", 1)
	rec := f.serve(t, http.HandlerFunc(f.ord.List), req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestOrderHandler_List_Error(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(errDBDown())

	req := f.authedRequest(t, http.MethodGet, "/api/user/orders", 1)
	rec := f.serve(t, http.HandlerFunc(f.ord.List), req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

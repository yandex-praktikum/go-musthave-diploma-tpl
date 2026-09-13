package service

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

func TestOrderService_UploadOrder_InvalidLuhn(t *testing.T) {
	svc, _ := newOrderFixture(t, NewAccrualClient("http://localhost"))

	if err := svc.UploadOrder(1, "not-luhn"); !errors.Is(err, domain.ErrInvalidOrderNumber) {
		t.Errorf("UploadOrder() error = %v, want ErrInvalidOrderNumber", err)
	}
}

func TestOrderService_UploadOrder_AlreadyUploadedByUser(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs("4561261212345467").
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("4561261212345467", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	if err := svc.UploadOrder(1, "4561261212345467"); !errors.Is(err, domain.ErrOrderAlreadyUploadedByUser) {
		t.Errorf("UploadOrder() error = %v, want ErrOrderAlreadyUploadedByUser", err)
	}
}

func TestOrderService_UploadOrder_AlreadyUploadedByAnother(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("4561261212345467", 2, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	if err := svc.UploadOrder(1, "4561261212345467"); !errors.Is(err, domain.ErrOrderAlreadyUploadedByAnotherUser) {
		t.Errorf("UploadOrder() error = %v, want ErrOrderAlreadyUploadedByAnotherUser", err)
	}
}

func TestOrderService_UploadOrder_Success(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs("4561261212345467").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.UploadOrder(1, "4561261212345467"); err != nil {
		t.Fatalf("UploadOrder() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestOrderService_UploadOrder_CheckError(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(errors.New("db down"))

	if err := svc.UploadOrder(1, "4561261212345467"); err == nil ||
		errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("UploadOrder() error = %v, want wrapped db error", err)
	}
}

func TestOrderService_UploadOrder_CreateError(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO orders`).WillReturnError(errors.New("db down"))

	if err := svc.UploadOrder(1, "4561261212345467"); err == nil {
		t.Error("UploadOrder() expected error, got nil")
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))
	now := time.Now()

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusProcessed, 100.0, now, now))

	orders, err := svc.GetUserOrders(1)
	if err != nil {
		t.Fatalf("GetUserOrders() error = %v", err)
	}
	if len(orders) != 1 || orders[0].Number != "123" {
		t.Errorf("unexpected orders: %+v", orders)
	}
}

func TestOrderService_GetUserOrders_Error(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(errors.New("db down"))

	if _, err := svc.GetUserOrders(1); err == nil {
		t.Error("GetUserOrders() expected error, got nil")
	}
}

// --- ProcessPendingOrders / processOrder ---

// newAccrualServer запускает тестовый HTTP-сервер accrual-системы.
func newAccrualServer(t *testing.T, status int, body string) (*httptest.Server, *AccrualClient) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
	t.Cleanup(server.Close)
	return server, NewAccrualClient(server.URL)
}

func TestOrderService_ProcessPendingOrders_Processed(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"PROCESSED","accrual":100.5}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE balances`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_Registered(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"REGISTERED"}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_Invalid(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"INVALID"}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_UnknownStatus(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"SOMETHING_ELSE"}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_AccrualError(t *testing.T) {
	// Сервер недоступен — ошибка не должна прерывать обработку.
	svc, mock := newOrderFixture(t, NewAccrualClient("http://127.0.0.1:1"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v, want nil", err)
	}
}

func TestOrderService_ProcessPendingOrders_UpdateError(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"PROCESSED","accrual":100.5}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).WillReturnError(errors.New("db down"))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v, want nil", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_AddBalanceError(t *testing.T) {
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"PROCESSED","accrual":100.5}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE balances`).WillReturnError(errors.New("db down"))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v, want nil", err)
	}
	server.Close()
}

func TestOrderService_ProcessPendingOrders_GetPendingError(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://localhost"))

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnError(errors.New("db down"))

	if err := svc.ProcessPendingOrders(10); err == nil {
		t.Error("ProcessPendingOrders() expected error, got nil")
	}
}

func TestOrderService_ProcessPendingOrders_ZeroAccrual(t *testing.T) {
	// PROCESSED с нулевым начислением — AddBalance не должен вызываться.
	server, client := newAccrualServer(t, http.StatusOK,
		`{"order":"123","status":"PROCESSED","accrual":0}`)
	svc, mock := newOrderFixture(t, client)

	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("123", 1, domain.OrderStatusNew, 0, time.Now(), time.Now()))
	mock.ExpectExec(`UPDATE orders`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.ProcessPendingOrders(10); err != nil {
		t.Fatalf("ProcessPendingOrders() error = %v", err)
	}
	server.Close()
}

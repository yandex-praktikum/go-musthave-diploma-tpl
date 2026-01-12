package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"gophermart/internal/repository"
	"gophermart/internal/service"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		userID         int64
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
	}{
		{
			name:        "new order accepted",
			orderNumber: "12345678903",
			userID:      1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT user_id FROM orders`).
					WithArgs("12345678903").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectExec(`INSERT INTO orders`).
					WithArgs(int64(1), "12345678903", "NEW", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantStatusCode: http.StatusAccepted,
		},
		{
			name:        "order already exists for same user",
			orderNumber: "12345678903",
			userID:      1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT user_id FROM orders`).
					WithArgs("12345678903").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "order exists for different user",
			orderNumber: "12345678903",
			userID:      1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT user_id FROM orders`).
					WithArgs("12345678903").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:        "invalid order number format",
			orderNumber: "123abc",
			userID:      1,
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:        "invalid luhn check",
			orderNumber: "12345678904",
			userID:      1,
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New() error = %v", err)
			}
			defer db.Close()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			orderRepo := repository.NewOrderRepository(db)
			orderService := service.NewOrderService(orderRepo)
			orderHandler := NewOrderHandler(orderService)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set("Content-Type", "text/plain")
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			orderHandler.CreateOrder(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("CreateOrder() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestOrderHandler_ListOrders(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
		wantOrders     int
	}{
		{
			name:   "successful list orders",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
					AddRow("12345678903", "PROCESSED", 100.5, time.Now()).
					AddRow("9278923470", "NEW", nil, time.Now())
				mock.ExpectQuery(`SELECT number, status, accrual, uploaded_at`).
					WithArgs(int64(1)).
					WillReturnRows(rows)
			},
			wantStatusCode: http.StatusOK,
			wantOrders:     2,
		},
		{
			name:   "no orders",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT number, status, accrual, uploaded_at`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}))
			},
			wantStatusCode: http.StatusNoContent,
			wantOrders:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New() error = %v", err)
			}
			defer db.Close()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			orderRepo := repository.NewOrderRepository(db)
			orderService := service.NewOrderService(orderRepo)
			orderHandler := NewOrderHandler(orderService)

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			orderHandler.ListOrders(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ListOrders() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantOrders > 0 && w.Code == http.StatusOK {
				var orders []map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&orders); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(orders) != tt.wantOrders {
					t.Errorf("ListOrders() orders count = %v, want %v", len(orders), tt.wantOrders)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

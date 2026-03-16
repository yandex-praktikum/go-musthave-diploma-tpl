package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"gophermart/internal/repository"
	"gophermart/internal/service"
)

func TestBalanceHandler_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
		wantBalance    float64
		wantWithdrawn  float64
	}{
		{
			name:   "successful balance retrieval",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(accrual\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(1000.5))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(200.0))
			},
			wantStatusCode: http.StatusOK,
			wantBalance:    800.5,
			wantWithdrawn:  200.0,
		},
		{
			name:   "zero balance",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(accrual\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
			},
			wantStatusCode: http.StatusOK,
			wantBalance:    0,
			wantWithdrawn:  0,
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

			balanceRepo := repository.NewBalanceRepository(db)
			withdrawalRepo := repository.NewWithdrawalRepository(db)
			orderRepo := repository.NewOrderRepository(db)
			balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, orderRepo, db)
			balanceHandler := NewBalanceHandler(balanceService)

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			balanceHandler.GetBalance(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("GetBalance() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if w.Code == http.StatusOK {
				var resp struct {
					Current   float64 `json:"current"`
					Withdrawn float64 `json:"withdrawn"`
				}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if resp.Current != tt.wantBalance {
					t.Errorf("GetBalance() current = %v, want %v", resp.Current, tt.wantBalance)
				}
				if resp.Withdrawn != tt.wantWithdrawn {
					t.Errorf("GetBalance() withdrawn = %v, want %v", resp.Withdrawn, tt.wantWithdrawn)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		body           map[string]interface{}
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
	}{
		{
			name:   "successful withdraw",
			userID: 1,
			body: map[string]interface{}{
				"order": "12345678903",
				"sum":   100.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(accrual\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(1000.0))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
				mock.ExpectExec(`INSERT INTO withdrawals`).
					WithArgs(int64(1), "12345678903", 100.0, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "insufficient funds",
			userID: 1,
			body: map[string]interface{}{
				"order": "12345678903",
				"sum":   1000.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(accrual\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(500.0))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(sum\), 0\)`).
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
				mock.ExpectRollback()
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
		{
			name:   "invalid order number",
			userID: 1,
			body: map[string]interface{}{
				"order": "123abc",
				"sum":   100.0,
			},
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

			balanceRepo := repository.NewBalanceRepository(db)
			withdrawalRepo := repository.NewWithdrawalRepository(db)
			orderRepo := repository.NewOrderRepository(db)
			balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, orderRepo, db)
			balanceHandler := NewBalanceHandler(balanceService)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			balanceHandler.Withdraw(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Withdraw() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

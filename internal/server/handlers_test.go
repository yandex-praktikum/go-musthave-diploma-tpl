package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"

	"gophermart/internal/config"
)

func TestServer_handleRegister(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
		wantCookie     bool
	}{
		{
			name:   "successful registration",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "testuser",
				"password": "testpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("testuser", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantStatusCode: http.StatusOK,
			wantCookie:     true,
		},
		{
			name:   "duplicate login",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "existing",
				"password": "testpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("existing", sqlmock.AnyArg()).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			wantStatusCode: http.StatusConflict,
			wantCookie:     false,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			body:   "invalid json",
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "empty login",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "",
				"password": "testpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			wantStatusCode: http.StatusBadRequest,
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

			s := &Server{
				cfg: &config.Config{},
				db:  db,
				mux: http.NewServeMux(),
			}
			s.registerRoutes()

			var bodyBytes []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyBytes = []byte(str)
				} else {
					bodyBytes, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/user/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.handleRegister(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleRegister() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "user_id" && c.Value != "" {
						found = true
						break
					}
				}
				if !found {
					t.Error("handleRegister() expected cookie 'user_id' not found")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestServer_handleLogin(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupMock      func(mock sqlmock.Sqlmock)
		wantStatusCode int
		wantCookie     bool
	}{
		{
			name:   "successful login",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "testuser",
				"password": "testpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
				mock.ExpectQuery(`SELECT id, password_hash FROM users`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hash)))
			},
			wantStatusCode: http.StatusOK,
			wantCookie:     true,
		},
		{
			name:   "wrong password",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "testuser",
				"password": "wrongpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
				mock.ExpectQuery(`SELECT id, password_hash FROM users`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hash)))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCookie:     false,
		},
		{
			name:   "user not found",
			method: http.MethodPost,
			body: map[string]string{
				"login":    "nonexistent",
				"password": "testpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, password_hash FROM users`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCookie:     false,
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

			s := &Server{
				cfg: &config.Config{},
				db:  db,
				mux: http.NewServeMux(),
			}
			s.registerRoutes()

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/user/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.handleLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "user_id" && c.Value != "" {
						found = true
						break
					}
				}
				if !found {
					t.Error("handleLogin() expected cookie 'user_id' not found")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestServer_handleCreateOrder(t *testing.T) {
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

			s := &Server{
				cfg: &config.Config{},
				db:  db,
				mux: http.NewServeMux(),
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set("Content-Type", "text/plain")
			req.AddCookie(&http.Cookie{Name: "user_id", Value: "1"})
			w := httptest.NewRecorder()

			s.handleCreateOrder(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleCreateOrder() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestServer_handleBalance(t *testing.T) {
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

			s := &Server{
				cfg: &config.Config{},
				db:  db,
				mux: http.NewServeMux(),
			}

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			req.AddCookie(&http.Cookie{Name: "user_id", Value: "1"})
			w := httptest.NewRecorder()

			s.handleBalance(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleBalance() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if w.Code == http.StatusOK {
				var resp balanceResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if resp.Current != tt.wantBalance {
					t.Errorf("handleBalance() current = %v, want %v", resp.Current, tt.wantBalance)
				}
				if resp.Withdrawn != tt.wantWithdrawn {
					t.Errorf("handleBalance() withdrawn = %v, want %v", resp.Withdrawn, tt.wantWithdrawn)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

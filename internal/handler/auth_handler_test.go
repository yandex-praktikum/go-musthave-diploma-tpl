package handler

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

	"gophermart/internal/repository"
	"gophermart/internal/service"
)

func TestAuthHandler_Register(t *testing.T) {
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

			userRepo := repository.NewUserRepository(db)
			authService := service.NewAuthService(userRepo)
			jwtService := service.NewJWTService("test-secret-key")
			authHandler := NewAuthHandler(authService, jwtService)

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

			authHandler.Register(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "token" && c.Value != "" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Register() expected cookie 'token' not found")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
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

			userRepo := repository.NewUserRepository(db)
			authService := service.NewAuthService(userRepo)
			jwtService := service.NewJWTService("test-secret-key")
			authHandler := NewAuthHandler(authService, jwtService)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/user/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			authHandler.Login(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "token" && c.Value != "" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Login() expected cookie 'user_id' not found")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations not met: %v", err)
			}
		})
	}
}

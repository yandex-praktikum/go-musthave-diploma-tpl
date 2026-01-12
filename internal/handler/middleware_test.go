package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/service"
)

func TestWithAuth(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")

	tests := []struct {
		name           string
		token          string
		wantStatusCode int
	}{
		{
			name:           "valid token",
			token:          "",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "no cookie",
			token:          "",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid.token.here",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			token:          "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMiddleware := NewAuthMiddleware(jwtService)

			var token string
			if tt.name == "valid token" {
				var err error
				token, err = jwtService.GenerateToken(123)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
			} else if tt.token != "" {
				token = tt.token
			}

			handler := authMiddleware.WithAuth(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if token != "" {
				req.AddCookie(&http.Cookie{Name: "token", Value: token})
			}
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("WithAuth() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	authMiddleware := NewAuthMiddleware(jwtService)

	tests := []struct {
		name   string
		userID int64
		wantID int64
		wantOk bool
	}{
		{
			name:   "valid user id",
			userID: 123,
			wantID: 123,
			wantOk: true,
		},
		{
			name:   "zero user id",
			userID: 1,
			wantID: 1,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtService.GenerateToken(tt.userID)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{Name: "token", Value: token})

			handler := authMiddleware.WithAuth(func(w http.ResponseWriter, r *http.Request) {
				gotID, gotOk := getUserIDFromContext(r)
				if gotID != tt.wantID {
					t.Errorf("getUserIDFromContext() id = %v, want %v", gotID, tt.wantID)
				}
				if gotOk != tt.wantOk {
					t.Errorf("getUserIDFromContext() ok = %v, want %v", gotOk, tt.wantOk)
				}
				w.WriteHeader(http.StatusOK)
			})

			w := httptest.NewRecorder()
			handler(w, req)
		})
	}
}

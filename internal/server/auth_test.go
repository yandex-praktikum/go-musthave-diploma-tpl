package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/config"
)

func TestServer_withAuth(t *testing.T) {
	tests := []struct {
		name           string
		cookieValue    string
		wantStatusCode int
	}{
		{
			name:           "valid cookie",
			cookieValue:    "1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "no cookie",
			cookieValue:    "",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid cookie value",
			cookieValue:    "invalid",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "zero user id",
			cookieValue:    "0",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "negative user id",
			cookieValue:    "-1",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg: &config.Config{},
				mux: http.NewServeMux(),
			}

			handler := s.withAuth(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "user_id", Value: tt.cookieValue})
			}
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("withAuth() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestServer_currentUserID(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
		wantID int64
		wantOk bool
	}{
		{
			name:   "valid cookie",
			cookie: &http.Cookie{Name: "user_id", Value: "123"},
			wantID: 123,
			wantOk: true,
		},
		{
			name:   "no cookie",
			cookie: nil,
			wantID: 0,
			wantOk: false,
		},
		{
			name:   "invalid value",
			cookie: &http.Cookie{Name: "user_id", Value: "abc"},
			wantID: 0,
			wantOk: false,
		},
		{
			name:   "zero value",
			cookie: &http.Cookie{Name: "user_id", Value: "0"},
			wantID: 0,
			wantOk: false,
		},
		{
			name:   "negative value",
			cookie: &http.Cookie{Name: "user_id", Value: "-1"},
			wantID: 0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg: &config.Config{},
				mux: http.NewServeMux(),
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			gotID, gotOk := s.currentUserID(req)

			if gotID != tt.wantID {
				t.Errorf("currentUserID() id = %v, want %v", gotID, tt.wantID)
			}
			if gotOk != tt.wantOk {
				t.Errorf("currentUserID() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

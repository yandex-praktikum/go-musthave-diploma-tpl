package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_NoToken(t *testing.T) {
	m := NewAuthMiddleware("secret")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called without token")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidTokenHeader(t *testing.T) {
	m := NewAuthMiddleware("secret")

	token, err := m.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var gotUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetUserID(r)
		if !ok {
			t.Error("GetUserID() = false, want true")
		}
		gotUserID = id
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if gotUserID != 42 {
		t.Errorf("userID = %d, want 42", gotUserID)
	}
}

func TestAuthMiddleware_ValidTokenCookie(t *testing.T) {
	m := NewAuthMiddleware("secret")

	token, err := m.GenerateToken(7)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	var gotUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetUserID(r)
		if !ok {
			t.Error("GetUserID() = false, want true")
		}
		gotUserID = id
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if gotUserID != 7 {
		t.Errorf("userID = %d, want 7", gotUserID)
	}
}

func TestAuthMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	m := NewAuthMiddleware("secret")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with invalid header")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Token abc")
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	m := NewAuthMiddleware("secret")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with invalid token")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_TokenFromOtherSecret(t *testing.T) {
	other := NewAuthMiddleware("other-secret")
	token, _ := other.GenerateToken(1)

	m := NewAuthMiddleware("secret")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with foreign token")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	m.Auth(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetUserID_MissingContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if _, ok := GetUserID(req); ok {
		t.Error("GetUserID() = true, want false for empty context")
	}
}

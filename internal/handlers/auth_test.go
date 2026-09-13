package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	f := newFixture(t)

	f.expectExists("alice", false)
	f.expectInsertUser(1)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"alice","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Проверяем установку cookie с токеном.
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("token cookie should be set")
	}

	// Проверяем заголовок Authorization.
	authHeader := rec.Header().Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		t.Errorf("Authorization header = %q, want Bearer token", authHeader)
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{invalid`))
	rec := httptest.NewRecorder()

	f.auth.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_EmptyFields(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"","password":""}`))
	rec := httptest.NewRecorder()

	f.auth.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_LoginExists(t *testing.T) {
	f := newFixture(t)

	f.expectExists("alice", true)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"alice","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Register_DBError(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT EXISTS`).WillReturnError(errDBDown())

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"alice","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Register(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	f := newFixture(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	f.expectFindUser(1, "alice", string(hash))

	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"alice","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("Authorization") == "" {
		t.Error("Authorization header should be set")
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Login_EmptyFields(t *testing.T) {
	f := newFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"ghost","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	f := newFixture(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	f.expectFindUser(1, "alice", string(hash))

	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"alice","password":"wrong"}`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_DBError(t *testing.T) {
	f := newFixture(t)

	f.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(errDBDown())

	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"alice","password":"password"}`))
	rec := httptest.NewRecorder()

	f.auth.Login(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

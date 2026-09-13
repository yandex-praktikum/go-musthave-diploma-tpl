package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register_Success(t *testing.T) {
	svc, mock := newAuthFixture(t)

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user, err := svc.Register("alice", "password")
	requireNoError(t, err)

	if user.Login != "alice" {
		t.Errorf("user.Login = %q, want alice", user.Login)
	}
	if user.Password == "password" {
		t.Error("password should be hashed")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password")) != nil {
		t.Error("stored password hash should match original password")
	}
	if user.CreatedAt.After(time.Now()) {
		t.Error("CreatedAt should not be in the future")
	}

	requireMockExpectations(t, mock)
}

func TestAuthService_Register_LoginExists(t *testing.T) {
	svc, mock := newAuthFixture(t)

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	_, err := svc.Register("alice", "password")
	requireError(t, err, domain.ErrLoginAlreadyExists)

	requireMockExpectations(t, mock)
}

func TestAuthService_Register_ExistsError(t *testing.T) {
	svc, mock := newAuthFixture(t)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnError(dbErr)

	_, err := svc.Register("alice", "password")
	requireError(t, err, dbErr)
	requireNotError(t, err, domain.ErrLoginAlreadyExists)

	requireMockExpectations(t, mock)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	svc, mock := newAuthFixture(t)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT EXISTS`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`INSERT INTO users`).WillReturnError(dbErr)

	_, err := svc.Register("alice", "password")
	requireError(t, err, dbErr)

	requireMockExpectations(t, mock)
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, mock := newAuthFixture(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
			AddRow(1, "alice", string(hash), time.Now()))

	user, err := svc.Login("alice", "password")
	requireNoError(t, err)

	if user.ID != 1 || user.Login != "alice" {
		t.Errorf("unexpected user: %+v", user)
	}

	requireMockExpectations(t, mock)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, mock := newAuthFixture(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
			AddRow(1, "alice", string(hash), time.Now()))

	_, err := svc.Login("alice", "wrong")
	requireError(t, err, domain.ErrInvalidCredentials)

	requireMockExpectations(t, mock)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, mock := newAuthFixture(t)

	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login("ghost", "password")
	requireError(t, err, domain.ErrInvalidCredentials)

	requireMockExpectations(t, mock)
}

func TestAuthService_Login_DBError(t *testing.T) {
	svc, mock := newAuthFixture(t)

	dbErr := errors.New("db down")
	mock.ExpectQuery(`SELECT id, login, password, created_at FROM users`).
		WillReturnError(dbErr)

	_, err := svc.Login("alice", "password")
	requireError(t, err, dbErr)
	requireNotError(t, err, domain.ErrInvalidCredentials)

	requireMockExpectations(t, mock)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &models.User{
		ID:           1,
		Login:        "alice",
		PasswordHash: "hash",
		Active:       true,
		CreatedAt:    time.Now(),
	}
	repo.EXPECT().
		Create(ctx, "alice", gomock.Any()).
		Return(user, nil)

	got, err := svc.Register(ctx, "alice", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if got.ID != 1 || got.Login != "alice" {
		t.Errorf("got user %+v", got)
	}
}

func TestUserService_Register_DuplicateLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.EXPECT().
		Create(ctx, "bob", gomock.Any()).
		Return(nil, &repository.ErrDuplicateLogin{Login: "bob"})

	_, err := svc.Register(ctx, "bob", "pass")
	if err == nil {
		t.Fatal("expected ErrDuplicateLogin")
	}
	var dup *repository.ErrDuplicateLogin
	if !errors.As(err, &dup) {
		t.Fatalf("expected *repository.ErrDuplicateLogin, got %T: %v", err, err)
	}
	if dup.Login != "bob" {
		t.Errorf("Login: got %q", dup.Login)
	}
}

func TestUserService_Register_ValidationEmptyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewUserService(mock.NewMockUserRepository(ctrl))
	ctx := context.Background()

	_, err := svc.Register(ctx, "", "password")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestUserService_Register_ValidationEmptyPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewUserService(mock.NewMockUserRepository(ctrl))
	ctx := context.Background()

	_, err := svc.Register(ctx, "alice", "")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestUserService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	user := &models.User{
		ID:           1,
		Login:        "charlie",
		PasswordHash: string(hash),
		Active:       true,
		CreatedAt:    time.Now(),
	}
	repo.EXPECT().
		GetByLogin(ctx, "charlie").
		Return(user, nil)

	got, err := svc.Login(ctx, "charlie", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if got.Login != "charlie" {
		t.Errorf("Login: got %q", got.Login)
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetByLogin(ctx, "nonexistent").
		Return(nil, &repository.ErrUserNotFound{Login: "nonexistent"})

	_, err := svc.Login(ctx, "nonexistent", "pass")
	if err == nil {
		t.Fatal("expected ErrInvalidCredentials")
	}
	var inv *ErrInvalidCredentials
	if !errors.As(err, &inv) {
		t.Fatalf("expected *ErrInvalidCredentials, got %T: %v", err, err)
	}
	if inv.Login != "nonexistent" {
		t.Errorf("Login: got %q", inv.Login)
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &models.User{
		ID:           1,
		Login:        "dave",
		PasswordHash: "not-a-bcrypt-hash",
		Active:       true,
		CreatedAt:    time.Now(),
	}
	repo.EXPECT().
		GetByLogin(ctx, "dave").
		Return(user, nil)

	_, err := svc.Login(ctx, "dave", "correct")
	if err == nil {
		t.Fatal("expected ErrInvalidCredentials")
	}
	var inv *ErrInvalidCredentials
	if !errors.As(err, &inv) {
		t.Fatalf("expected *ErrInvalidCredentials, got %T: %v", err, err)
	}
	if inv.Login != "dave" {
		t.Errorf("Login: got %q", inv.Login)
	}
}

func TestUserService_Login_ValidationEmptyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewUserService(mock.NewMockUserRepository(ctrl))
	ctx := context.Background()

	_, err := svc.Login(ctx, "", "password")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestUserService_Login_ValidationEmptyPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewUserService(mock.NewMockUserRepository(ctrl))
	ctx := context.Background()

	_, err := svc.Login(ctx, "alice", "")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

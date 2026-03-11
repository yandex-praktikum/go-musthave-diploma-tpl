package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

func TestUserRepository_Create(t *testing.T) {
	repos := setupDB(t)
	repo := repos.User
	ctx := context.Background()

	user, err := repo.Create(ctx, "alice", "hash123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if user.Login != "alice" {
		t.Errorf("Login: got %q", user.Login)
	}
	if user.PasswordHash != "hash123" {
		t.Errorf("PasswordHash: got %q", user.PasswordHash)
	}
	if !user.Active {
		t.Error("expected Active true")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestUserRepository_Create_DuplicateLogin(t *testing.T) {
	repos := setupDB(t)
	repo := repos.User
	ctx := context.Background()

	_, err := repo.Create(ctx, "Max", "hash1")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err = repo.Create(ctx, "Max", "hash2")
	if err == nil {
		t.Fatal("expected ErrDuplicateLogin")
	}
	var dup *repository.ErrDuplicateLogin
	if !errors.As(err, &dup) {
		t.Fatalf("expected *ErrDuplicateLogin, got %T: %v", err, err)
	}
	if dup.Login != "Max" {
		t.Errorf("ErrDuplicateLogin.Login: got %q", dup.Login)
	}
}

func TestUserRepository_GetByLogin_Found(t *testing.T) {
	repos := setupDB(t)
	repo := repos.User
	ctx := context.Background()

	created, err := repo.Create(ctx, "Max", "secret")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	user, err := repo.GetByLogin(ctx, "Max")
	if err != nil {
		t.Fatalf("GetByLogin: %v", err)
	}
	if user.ID != created.ID {
		t.Errorf("ID: got %d", user.ID)
	}
	if user.Login != "Max" {
		t.Errorf("Login: got %q", user.Login)
	}
	if user.PasswordHash != "secret" {
		t.Errorf("PasswordHash: got %q", user.PasswordHash)
	}
}

func TestUserRepository_GetByLogin_NotFound(t *testing.T) {
	repos := setupDB(t)
	repo := repos.User
	ctx := context.Background()

	_, err := repo.GetByLogin(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected ErrUserNotFound")
	}
	var notFound *repository.ErrUserNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *ErrUserNotFound, got %T: %v", err, err)
	}
}

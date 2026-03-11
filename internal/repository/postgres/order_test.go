package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

func TestOrderRepository_Create(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash123")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	order, err := orderRepo.Create(ctx, user.ID, "12345678903", "NEW")
	if err != nil {
		t.Fatalf("Create order: %v", err)
	}

	if order.UserID != user.ID {
		t.Errorf("UserID: got %d", order.UserID)
	}
	if order.Number != "12345678903" {
		t.Errorf("Number: got %q", order.Number)
	}
	if order.Status != "NEW" {
		t.Errorf("Status: got %q", order.Status)
	}
	if order.Accrual != nil {
		t.Errorf("expected Accrual nil for new order, got %d", *order.Accrual)
	}
	if order.UploadedAt.IsZero() {
		t.Error("expected non-zero UploadedAt")
	}
}

func TestOrderRepository_GetByNumber_Found(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "bob", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	created, err := orderRepo.Create(ctx, user.ID, "9278923470", "PROCESSED")
	if err != nil {
		t.Fatalf("Create order: %v", err)
	}

	order, err := orderRepo.GetByNumber(ctx, "9278923470")
	if err != nil {
		t.Fatalf("GetByNumber: %v", err)
	}
	if order.ID != created.ID {
		t.Errorf("ID: got %d", order.ID)
	}
	if order.UserID != user.ID {
		t.Errorf("UserID: got %d", order.UserID)
	}
	if order.Number != "9278923470" {
		t.Errorf("Number: got %q", order.Number)
	}
	if order.Status != "PROCESSED" {
		t.Errorf("Status: got %q", order.Status)
	}
}

func TestOrderRepository_GetByNumber_NotFound(t *testing.T) {
	repos := setupDB(t)
	orderRepo := repos.Order
	ctx := context.Background()

	_, err := orderRepo.GetByNumber(ctx, "99999999999")
	if err == nil {
		t.Fatal("expected ErrOrderNotFound")
	}
	var notFound *repository.ErrOrderNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *ErrOrderNotFound, got %T: %v", err, err)
	}
}

func TestOrderRepository_Create_DuplicateNumberFails(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_, err = orderRepo.Create(ctx, user.ID, "12345678903", "NEW")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err = orderRepo.Create(ctx, user.ID, "12345678903", "NEW")
	if err == nil {
		t.Fatal("expected error on duplicate number")
	}
}

func TestOrderRepository_ListByUserID(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_, err = orderRepo.Create(ctx, user.ID, "111", "NEW")
	if err != nil {
		t.Fatalf("Create order 1: %v", err)
	}
	_, err = orderRepo.Create(ctx, user.ID, "222", "PROCESSING")
	if err != nil {
		t.Fatalf("Create order 2: %v", err)
	}

	list, err := orderRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(list))
	}
	if list[0].Number != "222" {
		t.Errorf("first order Number: got %q", list[0].Number)
	}
	if list[1].Number != "111" {
		t.Errorf("second order Number: got %q", list[1].Number)
	}
}

func TestOrderRepository_ListByUserID_Empty(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	list, err := orderRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestOrderRepository_ListByUserID_OnlyThisUser(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user1, _ := userRepo.Create(ctx, "alice", "hash1")
	user2, _ := userRepo.Create(ctx, "bob", "hash2")
	_, _ = orderRepo.Create(ctx, user1.ID, "111", "NEW")
	_, _ = orderRepo.Create(ctx, user2.ID, "222", "NEW")

	list, err := orderRepo.ListByUserID(ctx, user1.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 order for user1, got %d", len(list))
	}
	if list[0].Number != "111" {
		t.Errorf("Number: got %q", list[0].Number)
	}
}

func TestOrderRepository_UpdateAccrualAndStatus_Updated(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_, err = orderRepo.Create(ctx, user.ID, "12345678903", "NEW")
	if err != nil {
		t.Fatalf("Create order: %v", err)
	}

	accrual := 500
	err = orderRepo.UpdateAccrualAndStatus(ctx, "12345678903", "PROCESSED", &accrual)
	if err != nil {
		t.Fatalf("UpdateAccrualAndStatus: %v", err)
	}

	order, err := orderRepo.GetByNumber(ctx, "12345678903")
	if err != nil {
		t.Fatalf("GetByNumber: %v", err)
	}
	if order.Status != "PROCESSED" {
		t.Errorf("Status: got %q", order.Status)
	}
	if order.Accrual == nil || *order.Accrual != 500 {
		t.Errorf("Accrual: got %v", order.Accrual)
	}
}

func TestOrderRepository_UpdateAccrualAndStatus_NotFound(t *testing.T) {
	repos := setupDB(t)
	orderRepo := repos.Order
	ctx := context.Background()

	err := orderRepo.UpdateAccrualAndStatus(ctx, "99999999999", "PROCESSED", nil)
	if err == nil {
		t.Fatal("expected ErrOrderNotFound")
	}
	var notFound *repository.ErrOrderNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *ErrOrderNotFound, got %T: %v", err, err)
	}
}

func TestOrderRepository_ListNumbersPendingAccrual(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_, _ = orderRepo.Create(ctx, user.ID, "111", "NEW")
	_, _ = orderRepo.Create(ctx, user.ID, "222", "PROCESSING")
	_, _ = orderRepo.Create(ctx, user.ID, "333", "PROCESSED")
	_, _ = orderRepo.Create(ctx, user.ID, "444", "INVALID")

	numbers, err := orderRepo.ListNumbersPendingAccrual(ctx, []string{"NEW", "PROCESSING"})
	if err != nil {
		t.Fatalf("ListNumbersPendingAccrual: %v", err)
	}
	if len(numbers) != 2 {
		t.Fatalf("expected 2 numbers (NEW, PROCESSING), got %d: %v", len(numbers), numbers)
	}
	if numbers[0] != "111" || numbers[1] != "222" {
		t.Errorf("expected [111, 222] by uploaded_at ASC, got %v", numbers)
	}
}

func TestOrderRepository_ListNumbersPendingAccrual_Empty(t *testing.T) {
	repos := setupDB(t)
	orderRepo, userRepo := repos.Order, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_, _ = orderRepo.Create(ctx, user.ID, "111", "PROCESSED")
	_, _ = orderRepo.Create(ctx, user.ID, "222", "INVALID")

	numbers, err := orderRepo.ListNumbersPendingAccrual(ctx, []string{"NEW", "PROCESSING"})
	if err != nil {
		t.Fatalf("ListNumbersPendingAccrual: %v", err)
	}
	if len(numbers) != 0 {
		t.Errorf("expected empty list, got %v", numbers)
	}
}

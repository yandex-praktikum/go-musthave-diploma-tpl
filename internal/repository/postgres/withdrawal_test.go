package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

func TestWithdrawalRepository_Create(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo := repos.Withdrawal, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	err = withdrawalRepo.Create(ctx, user.ID, "2377225624", 500)
	if err != nil {
		t.Fatalf("Create withdrawal: %v", err)
	}

	total, err := withdrawalRepo.GetTotalWithdrawnByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetTotalWithdrawnByUserID: %v", err)
	}
	if total != 500 {
		t.Errorf("GetTotalWithdrawnByUserID: got %d", total)
	}
}

func TestWithdrawalRepository_Create_DuplicateOrderFails(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo := repos.Withdrawal, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	err = withdrawalRepo.Create(ctx, user.ID, "2377225624", 500)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}

	err = withdrawalRepo.Create(ctx, user.ID, "2377225624", 100)
	if err == nil {
		t.Fatal("expected error on duplicate order")
	}
	var dup *repository.ErrDuplicateWithdrawalOrder
	if !errors.As(err, &dup) {
		t.Fatalf("expected *ErrDuplicateWithdrawalOrder, got %T: %v", err, err)
	}
	if dup.Order != "2377225624" {
		t.Errorf("Order: got %q", dup.Order)
	}
}

func TestWithdrawalRepository_GetTotalWithdrawnByUserID_Empty(t *testing.T) {
	repos := setupDB(t)
	userRepo := repos.User
	withdrawalRepo := repos.Withdrawal
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	total, err := withdrawalRepo.GetTotalWithdrawnByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetTotalWithdrawnByUserID: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0, got %d", total)
	}
}

func TestWithdrawalRepository_GetTotalWithdrawnByUserID_Sum(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo := repos.Withdrawal, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_ = withdrawalRepo.Create(ctx, user.ID, "111", 100)
	_ = withdrawalRepo.Create(ctx, user.ID, "222", 250)

	total, err := withdrawalRepo.GetTotalWithdrawnByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetTotalWithdrawnByUserID: %v", err)
	}
	if total != 350 {
		t.Errorf("expected 350, got %d", total)
	}
}

func TestWithdrawalRepository_ListByUserID_Empty(t *testing.T) {
	repos := setupDB(t)
	userRepo := repos.User
	withdrawalRepo := repos.Withdrawal
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	list, err := withdrawalRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestWithdrawalRepository_ListByUserID_OrderDesc(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo := repos.Withdrawal, repos.User
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	_ = withdrawalRepo.Create(ctx, user.ID, "first", 100)
	_ = withdrawalRepo.Create(ctx, user.ID, "second", 200)
	_ = withdrawalRepo.Create(ctx, user.ID, "third", 300)

	list, err := withdrawalRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}
	// newest first: third, second, first
	if list[0].Order != "third" || list[0].Sum != 300 {
		t.Errorf("first item: got order %q sum %d", list[0].Order, list[0].Sum)
	}
	if list[1].Order != "second" || list[1].Sum != 200 {
		t.Errorf("second item: got order %q sum %d", list[1].Order, list[1].Sum)
	}
	if list[2].Order != "first" || list[2].Sum != 100 {
		t.Errorf("third item: got order %q sum %d", list[2].Order, list[2].Sum)
	}
	if list[0].ProcessedAt.Before(list[1].ProcessedAt) || list[1].ProcessedAt.Before(list[2].ProcessedAt) {
		t.Error("expected processed_at DESC (newest first)")
	}
}

func TestWithdrawalRepository_ListByUserID_OnlyThisUser(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo := repos.Withdrawal, repos.User
	ctx := context.Background()

	user1, _ := userRepo.Create(ctx, "alice", "hash1")
	user2, _ := userRepo.Create(ctx, "bob", "hash2")
	_ = withdrawalRepo.Create(ctx, user1.ID, "111", 100)
	_ = withdrawalRepo.Create(ctx, user2.ID, "222", 200)

	list, err := withdrawalRepo.ListByUserID(ctx, user1.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 for user1, got %d", len(list))
	}
	if list[0].Order != "111" || list[0].Sum != 100 {
		t.Errorf("got order %q sum %d", list[0].Order, list[0].Sum)
	}
}

// TestWithdrawalRepository_Withdraw_Concurrent использует конкурентные вызовы для проверки
// атомарности операции Withdraw с advisory lock. Только один спрос должен пройти успешно.
func TestWithdrawalRepository_Withdraw_Concurrent(t *testing.T) {
	repos := setupDB(t)
	withdrawalRepo, userRepo, orderRepo := repos.Withdrawal, repos.User, repos.Order
	ctx := context.Background()

	// Создаём пользователя с балансом 100 (один заказ PROCESSED с accrual=100)
	user, err := userRepo.Create(ctx, "alice", "hash")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	order, err := orderRepo.Create(ctx, user.ID, "79927398713", "PROCESSED")
	if err != nil {
		t.Fatalf("Create order: %v", err)
	}
	// Обновляем accrual напрямую через SQL, так как нужно установить accrual для теста
	accrual := 100
	err = orderRepo.UpdateAccrualAndStatus(ctx, order.Number, "PROCESSED", &accrual)
	if err != nil {
		t.Fatalf("UpdateAccrualAndStatus: %v", err)
	}

	// Запускаем 10 конкурентных попыток списать по 100 (баланс = 100, только одна должна пройти)
	const numGoroutines = 10
	const withdrawAmount = int64(100)

	var wg sync.WaitGroup
	successCount := 0
	var successMu sync.Mutex
	errs := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			orderNum := "2377225624" + string(rune('0'+idx))
			err := withdrawalRepo.Withdraw(ctx, user.ID, orderNum, withdrawAmount)
			errs[idx] = err
			if err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	// Проверяем, что только один спрос прошёл успешно
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful withdrawal, got %d", successCount)
	}

	// Проверяем, что остальные получили ErrInsufficientFunds
	insufficientCount := 0
	for _, err := range errs {
		if err != nil {
			var insufficient *repository.ErrInsufficientFunds
			if errors.As(err, &insufficient) {
				insufficientCount++
			}
		}
	}
	if insufficientCount != numGoroutines-1 {
		t.Errorf("expected %d ErrInsufficientFunds errors, got %d", numGoroutines-1, insufficientCount)
	}

	// Проверяем, что в БД ровно одна запись о списании
	total, err := withdrawalRepo.GetTotalWithdrawnByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetTotalWithdrawnByUserID: %v", err)
	}
	if total != withdrawAmount {
		t.Errorf("expected total withdrawn %d, got %d", withdrawAmount, total)
	}
}

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
)

func TestBalanceService_GetBalance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderRepo := mock.NewMockOrderRepository(ctrl)
	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(orderRepo, withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)

	orderRepo.EXPECT().
		GetTotalAccrualsByUserID(ctx, userID).
		Return(int64(500), nil)
	withdrawalRepo.EXPECT().
		GetTotalWithdrawnByUserID(ctx, userID).
		Return(int64(100), nil)

	got, err := svc.GetBalance(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if got == nil {
		t.Fatal("got nil Balance")
	}
	if got.Current != 400 {
		t.Errorf("Current: got %d, want 400", got.Current)
	}
	if got.Withdrawn != 100 {
		t.Errorf("Withdrawn: got %d, want 100", got.Withdrawn)
	}
}

func TestBalanceService_GetBalance_OrderRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderRepo := mock.NewMockOrderRepository(ctrl)
	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(orderRepo, withdrawalRepo)
	ctx := context.Background()

	orderRepo.EXPECT().
		GetTotalAccrualsByUserID(ctx, int64(1)).
		Return(int64(0), errors.New("db error"))

	_, err := svc.GetBalance(ctx, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBalanceService_GetBalance_WithdrawalRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderRepo := mock.NewMockOrderRepository(ctrl)
	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(orderRepo, withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)

	orderRepo.EXPECT().
		GetTotalAccrualsByUserID(ctx, userID).
		Return(int64(100), nil)
	withdrawalRepo.EXPECT().
		GetTotalWithdrawnByUserID(ctx, userID).
		Return(int64(0), errors.New("db error"))

	_, err := svc.GetBalance(ctx, userID)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBalanceService_Withdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderRepo := mock.NewMockOrderRepository(ctrl)
	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(orderRepo, withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)
	sum := int64(50)

	withdrawalRepo.EXPECT().
		Withdraw(ctx, userID, validLuhnNumber, sum).
		Return(nil)

	err := svc.Withdraw(ctx, userID, validLuhnNumber, sum)
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
}

func TestBalanceService_Withdraw_TrimSpace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)
	sum := int64(10)

	withdrawalRepo.EXPECT().
		Withdraw(ctx, userID, validLuhnNumber, sum).
		Return(nil)

	err := svc.Withdraw(ctx, userID, "  "+validLuhnNumber+"  ", sum)
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
}

func TestBalanceService_Withdraw_InvalidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()

	// Withdraw не должен вызывать репозиторий при невалидном номере
	err := svc.Withdraw(ctx, 1, "1234567890", 100)
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestBalanceService_Withdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)
	orderNum := validLuhnNumber
	sum := int64(100)

	withdrawalRepo.EXPECT().
		Withdraw(ctx, userID, orderNum, sum).
		Return(&repository.ErrInsufficientFunds{Order: orderNum})

	err := svc.Withdraw(ctx, userID, orderNum, sum)
	if err == nil {
		t.Fatal("expected ErrInsufficientFunds")
	}
	var insufficient *repository.ErrInsufficientFunds
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected *repository.ErrInsufficientFunds, got %T: %v", err, err)
	}
	if insufficient.Order != orderNum {
		t.Errorf("Order: got %q", insufficient.Order)
	}
}

func TestBalanceService_Withdraw_DuplicateWithdrawalOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)
	orderNum := validLuhnNumber
	sum := int64(50)

	withdrawalRepo.EXPECT().
		Withdraw(ctx, userID, orderNum, sum).
		Return(&repository.ErrDuplicateWithdrawalOrder{Order: orderNum})

	err := svc.Withdraw(ctx, userID, orderNum, sum)
	if err == nil {
		t.Fatal("expected ErrDuplicateWithdrawalOrder")
	}
	var dup *repository.ErrDuplicateWithdrawalOrder
	if !errors.As(err, &dup) {
		t.Fatalf("expected *repository.ErrDuplicateWithdrawalOrder, got %T: %v", err, err)
	}
	if dup.Order != orderNum {
		t.Errorf("Order: got %q", dup.Order)
	}
}

func TestBalanceService_Withdraw_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()

	withdrawalRepo.EXPECT().
		Withdraw(ctx, int64(1), validLuhnNumber, int64(10)).
		Return(errors.New("connection lost"))

	err := svc.Withdraw(ctx, 1, validLuhnNumber, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBalanceService_ListWithdrawals_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()
	userID := int64(1)

	want := []*models.Withdrawal{
		{ID: 1, UserID: userID, Order: "79927398713", Sum: 100, ProcessedAt: time.Now()},
	}
	withdrawalRepo.EXPECT().
		ListByUserID(ctx, userID).
		Return(want, nil)

	got, err := svc.ListWithdrawals(ctx, userID)
	if err != nil {
		t.Fatalf("ListWithdrawals: %v", err)
	}
	if len(got) != 1 || got[0].Order != "79927398713" || got[0].Sum != 100 {
		t.Errorf("got %+v", got)
	}
}

func TestBalanceService_ListWithdrawals_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()

	withdrawalRepo.EXPECT().
		ListByUserID(ctx, int64(1)).
		Return(nil, nil)

	got, err := svc.ListWithdrawals(ctx, 1)
	if err != nil {
		t.Fatalf("ListWithdrawals: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil list, got %v", got)
	}
}

func TestBalanceService_ListWithdrawals_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawalRepo := mock.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(mock.NewMockOrderRepository(ctrl), withdrawalRepo)
	ctx := context.Background()

	withdrawalRepo.EXPECT().
		ListByUserID(ctx, int64(1)).
		Return(nil, errors.New("db error"))

	_, err := svc.ListWithdrawals(ctx, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Атомарность Withdraw (advisory lock + транзакция) тестируется не в сервисном слое:
// сервис только вызывает withdrawalRepo.Withdraw. Проверять отсутствие lost update нужно
// в интеграционных тестах репозитория (internal/repository/postgres):
//
// 1. Поднять реальную БД (testcontainers или CI PostgreSQL).
// 2. Создать одного пользователя с балансом 100 (один заказ PROCESSED с accrual=100).
// 3. Запустить N горутин (например 10), каждая вызывает Withdraw(ctx, userID, order_i, 100).
//    Только один спрос должен пройти (сумма 100, списать можно только один раз).
// 4. Проверить: ровно одна запись в withdrawals с sum=100, остальные вызовы вернули
//    ErrInsufficientFunds. Без advisory lock часть горутин могла бы успешно списать (lost update).
//
// Альтернатива: тест в postgres/withdrawal_test.go с реальным пулом — два параллельных
// Withdraw по одному userID на сумму больше половины баланса; один успешен, второй — ErrInsufficientFunds.

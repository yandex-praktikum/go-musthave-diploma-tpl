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

// 79927398713 — валидный номер по алгоритму Луна.
const validLuhnNumber = "79927398713"

func TestOrderService_AddOrder_SuccessNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()
	userID := int64(1)

	repo.EXPECT().
		GetByNumber(ctx, validLuhnNumber).
		Return(nil, &repository.ErrOrderNotFound{Number: validLuhnNumber})

	order := &models.Order{
		ID:         1,
		UserID:     userID,
		Number:     validLuhnNumber,
		Status:     OrderStatusNew,
		UploadedAt: time.Now(),
	}
	repo.EXPECT().
		Create(ctx, userID, validLuhnNumber, OrderStatusNew).
		Return(order, nil)

	got, created, err := svc.AddOrder(ctx, userID, validLuhnNumber)
	if err != nil {
		t.Fatalf("AddOrder: %v", err)
	}
	if !created {
		t.Error("expected created true")
	}
	if got.ID != 1 || got.Number != validLuhnNumber || got.UserID != userID {
		t.Errorf("got order %+v", got)
	}
	if got.Status != OrderStatusNew {
		t.Errorf("Status: got %q", got.Status)
	}
}

func TestOrderService_AddOrder_AlreadyUploadedByThisUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()
	userID := int64(1)

	existing := &models.Order{
		ID:         1,
		UserID:     userID,
		Number:     validLuhnNumber,
		Status:     OrderStatusProcessing,
		UploadedAt: time.Now(),
	}
	repo.EXPECT().
		GetByNumber(ctx, validLuhnNumber).
		Return(existing, nil)

	got, created, err := svc.AddOrder(ctx, userID, validLuhnNumber)
	if err != nil {
		t.Fatalf("AddOrder: %v", err)
	}
	if created {
		t.Error("expected created false (already existed)")
	}
	if got.ID != 1 || got.UserID != userID {
		t.Errorf("got order %+v", got)
	}
}

func TestOrderService_AddOrder_AlreadyUploadedByOtherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()
	userID := int64(2)

	existing := &models.Order{
		ID:         1,
		UserID:     1, // другой пользователь
		Number:     validLuhnNumber,
		Status:     OrderStatusNew,
		UploadedAt: time.Now(),
	}
	repo.EXPECT().
		GetByNumber(ctx, validLuhnNumber).
		Return(existing, nil)

	got, created, err := svc.AddOrder(ctx, userID, validLuhnNumber)
	if err == nil {
		t.Fatal("expected ErrOrderOwnedByOther")
	}
	if got != nil || created {
		t.Errorf("got %v, created %v", got, created)
	}
	var other *ErrOrderOwnedByOther
	if !errors.As(err, &other) {
		t.Fatalf("expected *ErrOrderOwnedByOther, got %T: %v", err, err)
	}
	if other.Number != validLuhnNumber {
		t.Errorf("Number: got %q", other.Number)
	}
}

func TestOrderService_AddOrder_ValidationEmptyNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewOrderService(mock.NewMockOrderRepository(ctrl))
	ctx := context.Background()

	_, _, err := svc.AddOrder(ctx, 1, "")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestOrderService_AddOrder_ValidationNonDigits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewOrderService(mock.NewMockOrderRepository(ctrl))
	ctx := context.Background()

	_, _, err := svc.AddOrder(ctx, 1, "1234abc")
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestOrderService_AddOrder_ValidationLuhnInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewOrderService(mock.NewMockOrderRepository(ctrl))
	ctx := context.Background()

	_, _, err := svc.AddOrder(ctx, 1, "12345678901")
	if err == nil {
		t.Fatal("expected ErrValidation (Luhn)")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestOrderService_AddOrder_GetByNumberError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetByNumber(ctx, validLuhnNumber).
		Return(nil, errors.New("db error"))

	_, _, err := svc.AddOrder(ctx, 1, validLuhnNumber)
	if err == nil {
		t.Fatal("expected error")
	}
	var other *ErrOrderOwnedByOther
	if errors.As(err, &other) {
		t.Error("expected db error, not ErrOrderOwnedByOther")
	}
}

func TestOrderService_ListOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()
	userID := int64(1)

	list := []*models.Order{
		{ID: 1, UserID: userID, Number: "111", Status: OrderStatusProcessed, Accrual: intPtr(100), UploadedAt: time.Now()},
		{ID: 2, UserID: userID, Number: "222", Status: OrderStatusNew, UploadedAt: time.Now()},
	}
	repo.EXPECT().
		ListByUserID(ctx, userID).
		Return(list, nil)

	got, err := svc.ListOrders(ctx, userID)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(got))
	}
	if got[0].Number != "111" || got[1].Number != "222" {
		t.Errorf("got %q, %q", got[0].Number, got[1].Number)
	}
}

func TestOrderService_ListOrders_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()

	repo.EXPECT().
		ListByUserID(ctx, int64(1)).
		Return(nil, nil)

	got, err := svc.ListOrders(ctx, 1)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestOrderService_ListOrders_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()

	repo.EXPECT().
		ListByUserID(ctx, int64(1)).
		Return(nil, errors.New("db error"))

	_, err := svc.ListOrders(ctx, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOrderService_GetOrderNumbersPendingAccrual_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()

	repo.EXPECT().
		ListNumbersPendingAccrual(ctx, OrderStatusesPendingAccrual).
		Return([]string{"111", "222"}, nil)

	numbers, err := svc.GetOrderNumbersPendingAccrual(ctx)
	if err != nil {
		t.Fatalf("GetOrderNumbersPendingAccrual: %v", err)
	}
	if len(numbers) != 2 || numbers[0] != "111" || numbers[1] != "222" {
		t.Errorf("got %v", numbers)
	}
}

func TestOrderService_ApplyAccrualResult_Updated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()
	accrual := 500

	repo.EXPECT().
		UpdateAccrualAndStatus(ctx, "12345678903", OrderStatusProcessed, &accrual).
		Return(nil)

	err := svc.ApplyAccrualResult(ctx, "12345678903", OrderStatusProcessed, &accrual)
	if err != nil {
		t.Fatalf("ApplyAccrualResult: %v", err)
	}
}

func TestOrderService_ApplyAccrualResult_ValidationInvalidStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewOrderService(mock.NewMockOrderRepository(ctrl))
	ctx := context.Background()

	err := svc.ApplyAccrualResult(ctx, "12345678903", "UNKNOWN", nil)
	if err == nil {
		t.Fatal("expected ErrValidation")
	}
	var val *ErrValidation
	if !errors.As(err, &val) {
		t.Fatalf("expected *ErrValidation, got %T: %v", err, err)
	}
}

func TestOrderService_ApplyAccrualResult_NoUpdateWhenProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo)
	ctx := context.Background()

	// PROCESSING не финальный — UpdateAccrualAndStatus не вызывается, возвращаем nil
	err := svc.ApplyAccrualResult(ctx, "12345678903", OrderStatusProcessing, nil)
	if err != nil {
		t.Fatalf("ApplyAccrualResult: %v", err)
	}
}

func intPtr(n int) *int {
	return &n
}

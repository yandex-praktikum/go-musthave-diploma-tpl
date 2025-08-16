package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	servicesmocks "github.com/vglushak/go-musthave-diploma-tpl/internal/services/mocks"
	storagemocks "github.com/vglushak/go-musthave-diploma-tpl/internal/storage/mocks"
	"go.uber.org/zap"
)

func TestNewOrderProcessor(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	interval := 5 * time.Second

	processor := NewOrderProcessor(mockStorage, mockAccrualService, interval, 5, zap.NewNop())

	assert.NotNil(t, processor)
	assert.Equal(t, mockStorage, processor.storage)
	assert.Equal(t, mockAccrualService, processor.accrualService)
	assert.Equal(t, interval, processor.interval)
	assert.NotNil(t, processor.stopChan)
}

func TestOrderProcessor_StartStop(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	interval := 100 * time.Millisecond

	processor := NewOrderProcessor(mockStorage, mockAccrualService, interval, 5, zap.NewNop())

	processor.Start()

	time.Sleep(50 * time.Millisecond)

	processor.Stop()

	time.Sleep(50 * time.Millisecond)
}

func TestOrderProcessor_ProcessOrder_Success(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"
	accrualValue := 100.0
	userID := int64(1)

	// Настраиваем моки
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: &accrualValue,
	}, nil)

	// Добавляем ожидания для получения заказа и баланса
	mockStorage.EXPECT().GetOrderByNumber(ctx, orderNumber).Return(&models.Order{
		ID:     1,
		UserID: userID,
		Number: orderNumber,
		Status: "NEW",
	}, nil)

	mockStorage.EXPECT().GetBalance(ctx, userID).Return(&models.Balance{
		UserID:    userID,
		Current:   50.0,
		Withdrawn: 0.0,
	}, nil)

	// Мок для транзакционного обновления
	mockStorage.EXPECT().UpdateOrderStatusAndBalance(ctx, orderNumber, "PROCESSED", &accrualValue, userID, 150.0, 0.0).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_InvalidOrder(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"

	// Моки - заказ не найден
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(nil, nil)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, orderNumber, "INVALID", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_RateLimitError(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"

	// Моки - превышение лимита запросов
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(nil, assert.AnError)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, orderNumber, "INVALID", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_ActualRateLimitError(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"

	// Моки - реальная rate limit ошибка
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(nil, services.ErrRateLimitExceeded)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, services.ErrRateLimitExceeded))
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "UpdateOrderStatus")
}

func TestOrderProcessor_ProcessOrder_NoAccrual(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"

	// Моки - заказ обработан, но без начисления
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)

	mockStorage.EXPECT().UpdateOrderStatus(ctx, orderNumber, "PROCESSED", (*float64)(nil)).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrdersWithWorkers_RateLimit(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orders := []models.Order{
		{ID: 1, UserID: 1, Number: "12345678903", Status: "NEW"},
		{ID: 2, UserID: 2, Number: "12345678904", Status: "PROCESSING"},
		{ID: 3, UserID: 3, Number: "12345678905", Status: "NEW"},
	}

	// Первый заказ успешно обрабатывается
	mockAccrualService.EXPECT().GetOrderInfo(ctx, "12345678903").Return(&models.AccrualResponse{
		Order:   "12345678903",
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, "12345678903", "PROCESSED", (*float64)(nil)).Return(nil)

	// Второй заказ вызывает rate limit
	mockAccrualService.EXPECT().GetOrderInfo(ctx, "12345678904").Return(nil, services.ErrRateLimitExceeded)

	// Третий заказ не должен обрабатываться из-за rate limit
	mockAccrualService.EXPECT().GetOrderInfo(ctx, "12345678905").Return(&models.AccrualResponse{
		Order:   "12345678905",
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, "12345678905", "PROCESSED", (*float64)(nil)).Return(nil)

	// Вызываем ProcessOrdersWithWorkers напрямую
	processor.ProcessOrdersWithWorkers(ctx, orders)

	// Проверяем, что первый заказ был обработан
	mockAccrualService.AssertCalled(t, "GetOrderInfo", ctx, "12345678903")
	mockStorage.AssertCalled(t, "UpdateOrderStatus", ctx, "12345678903", "PROCESSED", (*float64)(nil))

	// Проверяем, что второй заказ вызвал rate limit
	mockAccrualService.AssertCalled(t, "GetOrderInfo", ctx, "12345678904")
}

func TestOrderProcessor_ProcessOrdersWithWorkers_Success(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orders := []models.Order{
		{ID: 1, UserID: 1, Number: "12345678903", Status: "NEW"},
		{ID: 2, UserID: 2, Number: "12345678904", Status: "PROCESSING"},
	}

	// Моки для успешной обработки всех заказов
	mockAccrualService.EXPECT().GetOrderInfo(ctx, "12345678903").Return(&models.AccrualResponse{
		Order:   "12345678903",
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, "12345678903", "PROCESSED", (*float64)(nil)).Return(nil)

	mockAccrualService.EXPECT().GetOrderInfo(ctx, "12345678904").Return(&models.AccrualResponse{
		Order:   "12345678904",
		Status:  "PROCESSED",
		Accrual: nil,
	}, nil)
	mockStorage.EXPECT().UpdateOrderStatus(ctx, "12345678904", "PROCESSED", (*float64)(nil)).Return(nil)

	// Вызываем ProcessOrdersWithWorkers напрямую
	processor.ProcessOrdersWithWorkers(ctx, orders)

	// Проверяем, что все заказы были обработаны
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_TransactionalUpdate(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"
	accrualValue := 100.0
	userID := int64(1)

	// Моки для успешной обработки с начислением
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: &accrualValue,
	}, nil)

	// Моки для получения заказа и баланса
	mockStorage.EXPECT().GetOrderByNumber(ctx, orderNumber).Return(&models.Order{
		ID:     1,
		UserID: userID,
		Number: orderNumber,
		Status: "NEW",
	}, nil)

	mockStorage.EXPECT().GetBalance(ctx, userID).Return(&models.Balance{
		UserID:    userID,
		Current:   50.0,
		Withdrawn: 0.0,
	}, nil)

	// Мок для транзакционного обновления
	mockStorage.EXPECT().UpdateOrderStatusAndBalance(ctx, orderNumber, "PROCESSED", &accrualValue, userID, 150.0, 0.0).Return(nil)

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.NoError(t, err)
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestOrderProcessor_ProcessOrder_TransactionalUpdate_Error(t *testing.T) {
	mockStorage := &storagemocks.Storage{}
	mockAccrualService := &servicesmocks.AccrualServiceIface{}
	processor := NewOrderProcessor(mockStorage, mockAccrualService, 5*time.Second, 5, zap.NewNop())

	ctx := context.Background()
	orderNumber := "12345678903"
	accrualValue := 100.0
	userID := int64(1)

	// Моки для успешной обработки с начислением
	mockAccrualService.EXPECT().GetOrderInfo(ctx, orderNumber).Return(&models.AccrualResponse{
		Order:   orderNumber,
		Status:  "PROCESSED",
		Accrual: &accrualValue,
	}, nil)

	// Моки для получения заказа и баланса
	mockStorage.EXPECT().GetOrderByNumber(ctx, orderNumber).Return(&models.Order{
		ID:     1,
		UserID: userID,
		Number: orderNumber,
		Status: "NEW",
	}, nil)

	mockStorage.EXPECT().GetBalance(ctx, userID).Return(&models.Balance{
		UserID:    userID,
		Current:   50.0,
		Withdrawn: 0.0,
	}, nil)

	// Мок для ошибки транзакционного обновления
	mockStorage.EXPECT().UpdateOrderStatusAndBalance(ctx, orderNumber, "PROCESSED", &accrualValue, userID, 150.0, 0.0).Return(fmt.Errorf("transaction failed"))

	err := processor.ProcessOrder(ctx, orderNumber)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update order status and balance transactionally")
	mockAccrualService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

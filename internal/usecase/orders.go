package usecase

import (
	"context"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"

	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
)

// OrderUseCase определяет бизнес-логику для работы с заказами
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error)
	ProcessOrderResult(ctx context.Context, result worker.PollResult) error
	LoadActiveOrdersToWorker(ctx context.Context) error
	StartOrderProcessing(ctx context.Context) error
	StopOrderProcessing(ctx context.Context) error
	GetProcessingStats(ctx context.Context) (map[entity.OrderStatus]int, error)
}

type orderUseCase struct {
	repo   repository.Repository
	worker worker.OrderWorker
}

// OrderAPI интерфейс для внешнего API заказов
type OrderAPI interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*orderclient.OrderResponse, error)
}

// NewOrderUseCase создает новый useCase для заказов
func NewOrderUseCase(
	repo repository.Repository,
	worker worker.OrderWorker,
) OrderUseCase {
	return &orderUseCase{
		repo:   repo,
		worker: worker,
	}
}

// CreateOrder создает новый заказ
func (uc *orderUseCase) CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error) {
	// Проверяем номер заказа с помощью алгоритма Луна
	if !utils.IsValidLuhn(orderNumber) {
		return nil, fmt.Errorf("invalid order number")
	}

	order, err := uc.repo.CreateOrder(ctx, userID, orderNumber, entity.OrderStatusNew)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Добавляем заказ в воркер для отслеживания
	if err := uc.worker.AddOrder(ctx, orderNumber, 0); err != nil {
		// Если не удалось добавить в воркер, все равно возвращаем созданный заказ
		// (он будет загружен при следующем запуске воркера)
		return order, nil
	}

	return order, nil
}

// GetUserOrders возвращает заказы пользователя
func (uc *orderUseCase) GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error) {
	orders, err := uc.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, nil
}

// GetOrderByNumber возвращает заказ по номеру
func (uc *orderUseCase) GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order, err := uc.repo.GetOrderByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

// ProcessOrderResult обрабатывает результат опроса заказа
func (uc *orderUseCase) ProcessOrderResult(ctx context.Context, result worker.PollResult) error {
	if result.Error != nil {

		return fmt.Errorf("poll error for order %s: %w", result.OrderNumber, result.Error)
	}

	if result.OrderInfo == nil {
		return fmt.Errorf("no order info for order %s", result.OrderNumber)
	}

	// Преобразуем статус из внешнего API в внутренний
	var status entity.OrderStatus
	switch result.OrderInfo.Status {
	case "REGISTERED", "PROCESSING":
		status = entity.OrderStatusProcessing
	case "INVALID":
		status = entity.OrderStatusInvalid
	case "PROCESSED":
		status = entity.OrderStatusProcessed
	default:
		status = entity.OrderStatusNew
	}

	// Обновляем статус заказа в БД
	err := uc.repo.UpdateStatus(ctx, result.OrderNumber, status, result.OrderInfo.Accrual)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Если заказ завершен, удаляем его из воркера
	if status == entity.OrderStatusInvalid || status == entity.OrderStatusProcessed {
		if err := uc.worker.RemoveOrder(result.OrderNumber); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			return nil
		}
	}

	return nil
}

// LoadActiveOrdersToWorker загружает активные заказы в воркер
func (uc *orderUseCase) LoadActiveOrdersToWorker(ctx context.Context) error {
	activeOrders, err := uc.repo.GetActiveOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	for _, order := range activeOrders {
		if err := uc.worker.AddOrder(ctx, order.Number, 0); err != nil {
			// Логируем ошибку, но продолжаем загрузку остальных заказов
			continue
		}
	}

	return nil
}

// StartOrderProcessing запускает обработку заказов
func (uc *orderUseCase) StartOrderProcessing(ctx context.Context) error {
	// Загружаем активные заказы в воркер
	if err := uc.LoadActiveOrdersToWorker(ctx); err != nil {
		return fmt.Errorf("failed to load active orders: %w", err)
	}

	// Запускаем воркер
	if err := uc.worker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	// Запускаем обработку результатов
	go uc.processWorkerResults(ctx)

	return nil
}

// StopOrderProcessing останавливает обработку заказов
func (uc *orderUseCase) StopOrderProcessing(ctx context.Context) error {
	uc.worker.Stop()
	return nil
}

// processWorkerResults обрабатывает результаты из воркера
func (uc *orderUseCase) processWorkerResults(ctx context.Context) {
	results := uc.worker.Results()

	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-results:
			if !ok {
				return
			}

			if err := uc.ProcessOrderResult(ctx, result); err != nil {
				// Логируем ошибку, но продолжаем обработку
				continue
			}
		}
	}
}

// GetProcessingStats возвращает статистику обработки заказов
func (uc *orderUseCase) GetProcessingStats(ctx context.Context) (map[entity.OrderStatus]int, error) {
	stats := uc.worker.GetOrderStats()

	// Получаем общую статистику из БД
	activeOrders, err := uc.repo.GetActiveOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}

	// Добавляем информацию о заказах, которые еще не загружены в воркер
	dbStats := make(map[entity.OrderStatus]int)
	for _, order := range activeOrders {
		dbStats[order.Status]++
	}

	// Объединяем статистику
	for status, count := range dbStats {
		if _, exists := stats[status]; !exists {
			stats[status] = 0
		}
		stats[status] += count
	}

	return stats, nil
}

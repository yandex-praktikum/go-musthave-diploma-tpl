package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
)

// orderUC реализация OrderUseCase
type orderUC struct {
	repo   *repository.Repository
	worker worker.OrderWorker

	// Для управления горутинами
	processingWg      sync.WaitGroup
	processingCancel  context.CancelFunc
	processingStopped chan struct{}
}

// NewOrderUsecase создает новый экземпляр orderUC
func NewOrderUsecase(
	repo *repository.Repository,
	worker worker.OrderWorker,
) OrderUseCase {
	return &orderUC{
		repo:              repo,
		worker:            worker,
		processingStopped: make(chan struct{}),
	}
}

// CreateOrder создает новый заказ
func (uc *orderUC) CreateOrder(ctx context.Context, userID int, orderNumber string) (*entity.Order, error) {
	// Проверяем номер заказа
	if !validateOrderNumber(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	// Проверяем существование заказа
	exists, err := uc.repo.Order().Exists(ctx, orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	if exists {
		// Получаем существующий заказ для проверки владельца
		existingOrder, err := uc.repo.Order().GetByNumber(ctx, orderNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing order: %w", err)
		}

		if existingOrder.UserID == userID {
			return existingOrder, nil
		}
		return nil, ErrOrderAlreadyExists
	}

	// Создаем новый заказ
	order, err := uc.repo.Order().Create(ctx, userID, orderNumber, entity.OrderStatusNew)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Добавляем заказ в worker для отслеживания
	if err := uc.worker.AddOrder(ctx, orderNumber, 0); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		// Заказ будет загружен при следующем запуске worker
	}

	return order, nil
}

// GetUserOrders возвращает заказы пользователя
func (uc *orderUC) GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error) {
	orders, err := uc.repo.Order().GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, nil
}

// GetOrderByNumber возвращает заказ по номеру
func (uc *orderUC) GetOrderByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order, err := uc.repo.Order().GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// ProcessOrderResult обрабатывает результат опроса заказа
func (uc *orderUC) ProcessOrderResult(ctx context.Context, result worker.PollResult) error {
	if result.Error != nil {
		return fmt.Errorf("poll error for order %s: %w", result.OrderNumber, result.Error)
	}

	if result.OrderInfo == nil {
		return fmt.Errorf("no order info for order %s", result.OrderNumber)
	}

	// Получаем заказ из базы
	order, err := uc.repo.Order().GetByNumber(ctx, result.OrderNumber)
	if err != nil {
		return fmt.Errorf("failed to get order %s: %w", result.OrderNumber, err)
	}

	// Определяем статус
	status := mapExternalStatusToInternal(entity.OrderStatus(result.OrderInfo.Status))

	// Обновляем заказ
	if result.OrderInfo.Accrual != nil {
		err = uc.repo.Order().UpdateAccrual(ctx, order.ID, *result.OrderInfo.Accrual, status)
	} else {
		err = uc.repo.Order().UpdateStatus(ctx, order.ID, status)
	}

	if err != nil {
		return fmt.Errorf("failed to update order %s: %w", result.OrderNumber, err)
	}

	// Если заказ завершен, удаляем его из worker
	if status == entity.OrderStatusInvalid || status == entity.OrderStatusProcessed {
		_ = uc.worker.RemoveOrder(result.OrderNumber)
	}

	return nil
}

// mapExternalStatusToInternal преобразует внешний статус во внутренний
func mapExternalStatusToInternal(externalStatus entity.OrderStatus) entity.OrderStatus {
	switch externalStatus {
	case "REGISTERED", "PROCESSING":
		return entity.OrderStatusProcessing
	case "INVALID":
		return entity.OrderStatusInvalid
	case "PROCESSED":
		return entity.OrderStatusProcessed
	default:
		return entity.OrderStatusNew
	}
}

// LoadActiveOrdersToWorker загружает активные заказы в worker
func (uc *orderUC) LoadActiveOrdersToWorker(ctx context.Context) error {
	activeOrders, err := uc.repo.Order().GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	for _, order := range activeOrders {
		if err := uc.worker.AddOrder(ctx, order.Number, 0); err != nil {
			// Логируем ошибку, но продолжаем загрузку
			continue
		}
	}

	return nil
}

// StartOrderProcessing запускает обработку заказов
func (uc *orderUC) StartOrderProcessing(ctx context.Context) error {
	// Создаем контекст с отменой для управления горутиной
	ctx, cancel := context.WithCancel(ctx)
	uc.processingCancel = cancel

	// Загружаем активные заказы в worker
	if err := uc.LoadActiveOrdersToWorker(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to load active orders: %w", err)
	}

	// Запускаем worker
	if err := uc.worker.Start(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start worker: %w", err)
	}

	// Запускаем обработку результатов в фоновом режиме
	uc.processingWg.Add(1)
	go uc.processWorkerResults(ctx)

	return nil
}

// StopOrderProcessing останавливает обработку заказов
func (uc *orderUC) StopOrderProcessing(ctx context.Context) error {
	// Отменяем контекст, чтобы завершить горутину
	if uc.processingCancel != nil {
		uc.processingCancel()
	}

	// Останавливаем worker
	uc.worker.Stop()

	// Ждем завершения горутины с таймаутом
	done := make(chan struct{})
	go func() {
		uc.processingWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Горутина завершилась корректно
		close(uc.processingStopped)
		return nil
	case <-time.After(5 * time.Second):
		// Таймаут ожидания
		return errors.New("timeout waiting for order processing to stop")
	}
}

// processWorkerResults обрабатывает результаты из worker
func (uc *orderUC) processWorkerResults(ctx context.Context) {
	defer uc.processingWg.Done()

	results := uc.worker.Results()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменен - выходим
			return
		case result, ok := <-results:
			if !ok {
				// Канал закрыт - выходим
				return
			}

			// Обрабатываем результат
			if err := uc.ProcessOrderResult(ctx, result); err != nil {
				// Логируем ошибку, но продолжаем обработку
			}
		}
	}
}

// GetProcessingStats возвращает статистику обработки заказов
func (uc *orderUC) GetProcessingStats(ctx context.Context) (map[entity.OrderStatus]int, error) {
	stats := uc.worker.GetOrderStats()

	// Получаем общую статистику из БД
	activeOrders, err := uc.repo.Order().GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}

	// Добавляем информацию о заказах, которые еще не загружены в worker
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

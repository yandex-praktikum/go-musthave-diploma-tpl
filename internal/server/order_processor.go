package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/services"
	"go.uber.org/zap"
)

// OrderProcessor обрабатывает заказы в фоновом режиме
type OrderProcessor struct {
	storage        Storage
	accrualService services.AccrualServiceIface
	interval       time.Duration
	stopChan       chan struct{}
	workerCount    int
	nextSendTime   atomic.Int64 // время следующей отправки в наносекундах
	logger         *zap.Logger
}

// NewOrderProcessor создает новый процессор заказов
func NewOrderProcessor(storage Storage, accrualService services.AccrualServiceIface, interval time.Duration, workerCount int, logger *zap.Logger) *OrderProcessor {
	return &OrderProcessor{
		storage:        storage,
		accrualService: accrualService,
		interval:       interval,
		stopChan:       make(chan struct{}),
		workerCount:    workerCount,
		logger:         logger,
	}
}

// Start запускает обработку заказов
func (p *OrderProcessor) Start() {
	go p.processLoop()
}

// Stop останавливает обработку заказов
func (p *OrderProcessor) Stop() {
	close(p.stopChan)
}

// processLoop основной цикл обработки
func (p *OrderProcessor) processLoop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Проверяем, не нужно ли подождать из-за rate limit
			nextSend := p.nextSendTime.Load()
			if nextSend > 0 {
				now := time.Now().UnixNano()
				if now < nextSend {
					waitTime := time.Duration(nextSend - now)
					p.logger.Info("Rate limit detected, waiting for cooldown", zap.Duration("cooldown", waitTime))

					timer := time.NewTimer(waitTime)
					select {
					case <-timer.C:
						// Время ожидания истекло
					case <-p.stopChan:
						timer.Stop()
						return
					}
				}
				// Сбрасываем время следующей отправки
				p.nextSendTime.Store(0)
			}
			p.ProcessOrders()
		case <-p.stopChan:
			return
		}
	}
}

// ProcessOrders обрабатывает заказы со статусом NEW и PROCESSING
func (p *OrderProcessor) ProcessOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const batchSize = 100
	offset := 0

	for {
		// Получаем заказы со статусом NEW и PROCESSING с пагинацией
		orders, err := p.storage.GetOrdersByStatusPaginated(ctx, []string{"NEW", "PROCESSING"}, batchSize, offset)
		if err != nil {
			p.logger.Error("Failed to get orders for processing", zap.Error(err))
			return
		}

		if len(orders) == 0 {
			break
		}

		p.logger.Info("Processing batch of orders",
			zap.Int("count", len(orders)),
			zap.Int("offset", offset),
			zap.Int("workers", p.workerCount))

		p.ProcessOrdersWithWorkers(ctx, orders)

		if len(orders) < batchSize {
			break
		}

		offset += batchSize
	}
}

// ProcessOrdersWithWorkers обрабатывает заказы параллельно
func (p *OrderProcessor) ProcessOrdersWithWorkers(ctx context.Context, orders []models.Order) {
	// Канал для передачи заказов воркерам
	orderChan := make(chan models.Order, len(orders))

	// WaitGroup для ожидания завершения всех воркеров
	var wg sync.WaitGroup

	// Запускаем воркеры
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go p.worker(ctx, orderChan, &wg)
	}

	// Отправляем заказы в канал
	go func() {
		defer close(orderChan)
		for _, order := range orders {
			select {
			case orderChan <- order:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

// worker обрабатывает заказы из канала
func (p *OrderProcessor) worker(ctx context.Context, orderChan <-chan models.Order, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case order, ok := <-orderChan:
			if !ok {
				return
			}

			// Проверяем, не нужно ли подождать из-за rate limit
			nextSend := p.nextSendTime.Load()
			if nextSend > 0 {
				now := time.Now().UnixNano()
				if now < nextSend {
					// Воркер завершает работу, так как есть rate limit
					return
				}
			}

			if err := p.ProcessOrder(ctx, order.Number); err != nil {
				var rateLimitErr *services.RateLimitError
				if errors.As(err, &rateLimitErr) {
					p.logger.Info("Rate limit exceeded, worker stopping",
						zap.Duration("retryAfter", rateLimitErr.RetryAfter))

					// Устанавливаем время следующей отправки
					nextSendTime := time.Now().Add(rateLimitErr.RetryAfter).UnixNano()
					p.nextSendTime.Store(nextSendTime)
					return
				}
				p.logger.Error("Failed to process order",
					zap.String("orderNumber", order.Number),
					zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

// ProcessOrder обрабатывает конкретный заказ
func (p *OrderProcessor) ProcessOrder(ctx context.Context, orderNumber string) error {
	// Получаем информацию о заказе из системы начисления
	accrualInfo, err := p.accrualService.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		// Если ошибка связана с превышением лимита запросов, не обновляем статус
		if errors.Is(err, services.ErrRateLimitExceeded) {
			return err
		}

		// Обновляем статус на INVALID
		return p.storage.UpdateOrderStatus(ctx, orderNumber, "INVALID", nil)
	}

	if accrualInfo == nil {
		// Заказ не найден в системе начисления
		return p.storage.UpdateOrderStatus(ctx, orderNumber, "INVALID", nil)
	}

	// Обновляем статус и начисление
	var accrual *float64
	if accrualInfo.Accrual != nil {
		accrual = accrualInfo.Accrual
	}

	// Если заказ обработан и есть начисление, обновляем статус и баланс в одной транзакции
	if accrualInfo.Status == "PROCESSED" && accrual != nil && *accrual > 0 {
		// Получаем заказ для определения пользователя
		order, err := p.storage.GetOrderByNumber(ctx, orderNumber)
		if err != nil {
			return fmt.Errorf("failed to get order for balance update: %w", err)
		}

		// Получаем текущий баланс
		balance, err := p.storage.GetBalance(ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		// Обновляем статус заказа и баланс пользователя атомарно
		newCurrent := balance.Current + *accrual
		if err := p.storage.UpdateOrderStatusAndBalance(ctx, orderNumber, accrualInfo.Status, accrual, order.UserID, newCurrent, balance.Withdrawn); err != nil {
			return fmt.Errorf("failed to update order status and balance transactionally: %w", err)
		}
		return nil
	}

	// В остальных случаях просто обновляем статус заказа
	if err := p.storage.UpdateOrderStatus(ctx, orderNumber, accrualInfo.Status, accrual); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

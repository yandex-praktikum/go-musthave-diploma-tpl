// Package worker предоставляет фоновый воркер для обработки заказов.
package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/services"
)

// Config конфигурация воркера
type Config struct {
	pollInterval time.Duration
	batchLimit   int
	retryDelay   time.Duration
}

// Option функциональная опция для Config
type Option func(*Config)

// WithPollInterval устанавливает интервал опроса заказов
func WithPollInterval(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.pollInterval = d
		}
	}
}

// WithBatchLimit устанавливает максимальное количество заказов за один цикл
func WithBatchLimit(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.batchLimit = n
		}
	}
}

// WithRetryDelay устанавливает задержку перед повторной попыткой при rate limit
func WithRetryDelay(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.retryDelay = d
		}
	}
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		pollInterval: time.Second * 1,
		batchLimit:   10,
		retryDelay:   time.Second * 60,
	}
}

// NewConfig создаёт конфигурацию с применением опций
func NewConfig(opts ...Option) Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// PollInterval возвращает интервал опроса заказов
func (c Config) PollInterval() time.Duration {
	return c.pollInterval
}

// BatchLimit возвращает максимальное количество заказов за один цикл
func (c Config) BatchLimit() int {
	return c.batchLimit
}

// RetryDelay возвращает задержку перед повторной попыткой при rate limit
func (c Config) RetryDelay() time.Duration {
	return c.retryDelay
}

// Worker фоновый воркер обработки заказов
type Worker struct {
	config        Config
	orderService  *services.OrderService
	accrualClient *accrual.Client
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// New создаёт новый воркер
func New(config Config, orderService *services.OrderService, accrualClient *accrual.Client) *Worker {
	return &Worker{
		config:        config,
		orderService:  orderService,
		accrualClient: accrualClient,
		stopChan:      make(chan struct{}),
	}
}

// Run запускает воркера в фоновом режиме
func (w *Worker) Run(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.run(ctx)
	}()
}

// run основной цикл воркера
func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PollInterval())
	defer ticker.Stop()

	logger.Info("Worker started", "poll_interval", w.config.PollInterval())

	// Запускаем обработку сразу при старте
	w.processOrders(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker: context canceled, stopping")
			return

		case <-w.stopChan:
			logger.Info("Worker: stop signal received")
			return

		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

// processOrders обрабатывает заказы
func (w *Worker) processOrders(ctx context.Context) {
	// Получаем заказы со статусом NEW и PROCESSING
	newOrders, err := w.orderService.GetOrdersByStatus(ctx, "NEW", w.config.BatchLimit())
	if err != nil {
		logger.Error("Worker: failed to get NEW orders", "error", err)
	} else {
		logger.Debug("Worker: processing NEW orders", "count", len(newOrders))
		for _, order := range newOrders {
			select {
			case <-ctx.Done():
				logger.Info("Worker: context canceled during NEW orders processing")
				return
			default:
				w.processOrder(ctx, order.ID, order.UserID, order.Number)
			}
		}
	}

	processingOrders, err := w.orderService.GetOrdersByStatus(ctx, "PROCESSING", w.config.BatchLimit())
	if err != nil {
		logger.Error("Worker: failed to get PROCESSING orders", "error", err)
	} else {
		logger.Debug("Worker: processing PROCESSING orders", "count", len(processingOrders))
		for _, order := range processingOrders {
			select {
			case <-ctx.Done():
				logger.Info("Worker: context canceled during PROCESSING orders processing")
				return
			default:
				w.processOrder(ctx, order.ID, order.UserID, order.Number)
			}
		}
	}
}

// processOrder обрабатывает один заказ
func (w *Worker) processOrder(ctx context.Context, orderID int64, userID int64, orderNumber string) {
	logger.Debug("Worker: processing order", "order", orderNumber, "id", orderID)

	// Создаём контекст с таймаутом для обработки одного заказа
	// Это гарантирует, что shutdown не будет ждать бесконечно
	orderCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	info, err := w.accrualClient.GetOrderInfo(orderCtx, orderNumber)
	if err != nil {
		// Проверяем на rate limit
		var rateLimitErr *accrual.RateLimitError
		if errors.As(err, &rateLimitErr) {
			// Rate limit - ждём и повторяем позже
			logger.Warn("Worker: rate limit", "order", orderNumber, "retry_after", rateLimitErr.RetryAfter)
			return
		}

		if errors.Is(err, accrual.ErrOrderNotFound) {
			// Заказ не найден в системе accrual - оставляем как NEW для повторной попытки
			logger.Debug("Worker: order not found in accrual system", "order", orderNumber)
			return
		}

		logger.Error("Worker: error getting order info", "order", orderNumber, "error", err)
		return
	}

	// Определяем статус для обновления
	var newStatus string
	switch info.Status {
	case accrual.StatusRegistered:
		newStatus = "PROCESSING"
	case accrual.StatusInvalid:
		newStatus = "INVALID"
	case accrual.StatusProcessing:
		newStatus = "PROCESSING"
	case accrual.StatusProcessed:
		newStatus = "PROCESSED"
	}

	// Если статус PROCESSED и есть начисление — начисляем баллы
	if info.Status == accrual.StatusProcessed && info.Accrual != nil && *info.Accrual > 0 {
		if err := w.orderService.ProcessOrderAccrual(orderCtx, orderID, userID, "PROCESSED", info.Accrual, orderNumber); err != nil {
			logger.Error("Worker: failed to process accrual", "order", orderNumber, "error", err)
			return
		}
		logger.Info("Worker: accrued points", "order", orderNumber, "accrual", *info.Accrual)
	} else {
		// Просто обновляем статус без начисления
		if err := w.orderService.UpdateOrderStatus(orderCtx, orderID, newStatus, info.Accrual); err != nil {
			logger.Error("Worker: failed to update order status", "order", orderNumber, "error", err)
			return
		}
	}

	logger.Info("Worker: order status updated", "order", orderNumber, "status", newStatus)
}

// Stop останавливает воркера
func (w *Worker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

// Shutdown gracefully останавливает воркера
func (w *Worker) Shutdown(ctx context.Context) error {
	close(w.stopChan)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Worker shutdown completed")
		return nil
	case <-ctx.Done():
		logger.Warn("Worker shutdown timed out")
		return ctx.Err()
	}
}

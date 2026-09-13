// internal/service/worker.go
package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

const (
	defaultInterval  = 5 * time.Second
	defaultBatchSize = 100
)

// Worker — фоновый воркер, периодически обрабатывающий заказы,
// ожидающие расчёта вознаграждения.
type Worker struct {
	orderService *OrderService
	ticker       *time.Ticker
	batchSize    int
	interval     time.Duration
}

// NewWorker создаёт новый Worker с интервалом опроса 5 секунд
// и размером пакета обработки 100 заказов.
func NewWorker(orderService *OrderService) *Worker {
	return &Worker{
		orderService: orderService,
		ticker:       time.NewTicker(defaultInterval),
		batchSize:    defaultBatchSize,
		interval:     defaultInterval,
	}
}

// Start запускает воркер.
//
// Воркер работает в цикле: обрабатывает пачку заказов, ждёт тик тикера,
// повторяет. Если accrual-сервис отвечает 429 Too Many Requests, воркер
// приостанавливается на время, указанное в Retry-After, слушая ctx.Done()
// для корректного завершения.
func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started, checking orders every", w.interval)

	// Сразу обрабатываем заказы при старте
	w.processOrders(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping...")
			w.ticker.Stop()
			return
		case <-w.ticker.C:
			w.processOrders(ctx)
		}
	}
}

// processOrders обрабатывает ожидающие заказы.
//
// При получении *domain.RateLimitError приостанавливает работу
// на указанное в ошибке время, слушая ctx.Done().
func (w *Worker) processOrders(ctx context.Context) {
	err := w.orderService.ProcessPendingOrders(w.batchSize)
	if err == nil {
		return
	}

	// Проверяем, не попросил ли accrual снизить нагрузку
	var rateLimitErr *domain.RateLimitError
	if errors.As(err, &rateLimitErr) {
		log.Printf("[WARN] Accrual rate limit hit, pausing for %s", rateLimitErr.RetryAfter)
		w.pause(ctx, rateLimitErr.RetryAfter)
		return
	}

	log.Printf("[ERROR] Worker error: %v", err)
}

// pause приостанавливает воркер на указанное время, слушая ctx.Done()
// для корректного завершения при остановке процесса.
func (w *Worker) pause(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		log.Println("Worker pause interrupted by context cancellation")
		return
	case <-timer.C:
		log.Println("Worker resumed after pause")
	}
}

// Stop останавливает воркер.
//
// Deprecated: используйте отмену ctx, переданного в Start.
func (w *Worker) Stop() {
	w.ticker.Stop()
}

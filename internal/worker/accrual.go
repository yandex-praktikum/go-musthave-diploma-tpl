package worker

import (
	"context"
	"errors"
	"time"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/accrual"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/logger"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	maxRetries429  = 3
	maxConcurrent  = 10
)

// OrderAccrualService — минимальный интерфейс сервиса заказов для воркера начислений.
type OrderAccrualService interface {
	GetOrderNumbersPendingAccrual(ctx context.Context) ([]string, error)
	ApplyAccrualResult(ctx context.Context, number, status string, accrual *int) error
}

// AccrualWorker — воркер опроса системы начислений и обновления заказов.
type AccrualWorker struct {
	baseURL  string
	orderSvc OrderAccrualService
	interval time.Duration
	client   accrual.Client
}

// NewAccrualWorker создаёт воркер. baseURL — адрес системы начислений (пустой = опрос отключён). interval — пауза между проходами.
func NewAccrualWorker(baseURL string, orderSvc OrderAccrualService, interval time.Duration) *AccrualWorker {
	return &AccrualWorker{baseURL: baseURL, orderSvc: orderSvc, interval: interval}
}

// Run запускает цикл опроса; блокируется до отмены ctx.
// Если baseURL пустой — логирует отключение и сразу выходит.
func (w *AccrualWorker) Run(ctx context.Context) {
	if w.baseURL == "" {
		logger.Log.Info("Accrual worker disabled: no accrual system address")
		return
	}
	w.client = accrual.NewHTTPClient(w.baseURL, nil)
	logger.Log.Info("Accrual worker started", zap.String("accrual_address", w.baseURL))
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		w.doOnePass(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// следующий проход
		}
	}
}

func (w *AccrualWorker) doOnePass(ctx context.Context) {
	numbers, err := w.orderSvc.GetOrderNumbersPendingAccrual(ctx)
	if err != nil || len(numbers) == 0 {
		return
	}
	
	// Используем errgroup для управления конкурентностью и сбора ошибок
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrent)
	
	for _, number := range numbers {
		num := number // capture loop variable
		g.Go(func() error {
			w.processOrder(gctx, num)
			return nil // processOrder логирует ошибки внутри, но не возвращает их
		})
	}
	
	if err := g.Wait(); err != nil {
		logger.Log.Error("Error processing orders in accrual worker", zap.Error(err))
	}
}

func (w *AccrualWorker) processOrder(ctx context.Context, number string) {
	for retry := 0; retry <= maxRetries429; retry++ {
		resp, err := w.client.GetOrder(ctx, number)
		if err == nil {
			if applyErr := w.orderSvc.ApplyAccrualResult(ctx, resp.Order, resp.Status, resp.Accrual); applyErr != nil {
				logger.Log.Error("Failed to apply accrual result",
					zap.String("order", resp.Order),
					zap.Error(applyErr))
			}
			return
		}
		var notReg *accrual.ErrOrderNotRegistered
		if errors.As(err, &notReg) {
			if applyErr := w.orderSvc.ApplyAccrualResult(ctx, number, service.OrderStatusInvalid, nil); applyErr != nil {
				logger.Log.Error("Failed to apply invalid status",
					zap.String("order", number),
					zap.Error(applyErr))
			}
			return
		}
		var rateLimit *accrual.ErrRateLimit
		if errors.As(err, &rateLimit) {
			if retry < maxRetries429 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(rateLimit.RetryAfter) * time.Second):
					continue
				}
			}
			logger.Log.Warn("Max retries reached for rate limit",
				zap.String("order", number),
				zap.Int("retries", retry))
			return
		}
		// 500 / сеть — пропускаем, в следующем цикле попадёт снова
		logger.Log.Debug("Temporary error processing order, will retry later",
			zap.String("order", number),
			zap.Error(err))
		return
	}
}

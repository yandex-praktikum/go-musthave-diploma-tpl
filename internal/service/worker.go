package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/true18/go-musthave-diploma-tpl/internal/accrual"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
)

const (
	defaultWorkerInterval = 2 * time.Second
	defaultWorkerBatch    = 20
)

// AccrualClient gets order states from the external accrual system.
type AccrualClient interface {
	Fetch(ctx context.Context, number string) (accrual.Result, error)
}

// Worker periodically polls pending orders and updates their accrual state.
type Worker struct {
	repo       Repository
	client     AccrualClient
	logger     *zap.Logger
	interval   time.Duration
	batchLimit int
	retryAfter time.Time
}

// NewWorker creates a Worker with conservative polling defaults.
func NewWorker(repo Repository, client AccrualClient, logger *zap.Logger) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Worker{
		repo:       repo,
		client:     client,
		logger:     logger,
		interval:   defaultWorkerInterval,
		batchLimit: defaultWorkerBatch,
	}
}

// Run starts polling until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	w.poll(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

// PollOnce processes one batch of pending orders. It is intended for tests.
func (w *Worker) PollOnce(ctx context.Context) {
	w.poll(ctx)
}

func (w *Worker) poll(ctx context.Context) {
	if w.client == nil || time.Now().Before(w.retryAfter) {
		return
	}

	orders, err := w.repo.ListPendingOrders(ctx, w.batchLimit)
	if err != nil {
		w.logger.Error("list pending orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		result, err := w.client.Fetch(ctx, order.Number)
		if errors.Is(err, accrual.ErrRateLimited) {
			w.retryAfter = time.Now().Add(result.RetryAfter)
			return
		}
		if err != nil {
			w.logger.Warn("fetch accrual", zap.String("order", order.Number), zap.Error(err))
			continue
		}
		if result.NoContent {
			continue
		}

		status, accrualAmount, ok := mapAccrualResult(result)
		if !ok {
			w.logger.Warn("unknown accrual status", zap.String("status", result.Status))
			continue
		}
		if err := w.repo.UpdateOrder(ctx, order.Number, status, accrualAmount); err != nil {
			w.logger.Warn("update order from accrual", zap.String("order", order.Number), zap.Error(err))
		}
	}
}

func mapAccrualResult(result accrual.Result) (string, *money.Amount, bool) {
	switch result.Status {
	case accrual.StatusRegistered, accrual.StatusProcessing:
		return repository.StatusProcessing, nil, true
	case accrual.StatusInvalid:
		return repository.StatusInvalid, nil, true
	case accrual.StatusProcessed:
		return repository.StatusProcessed, result.Accrual, true
	default:
		return "", nil, false
	}
}

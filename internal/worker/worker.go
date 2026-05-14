package worker

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/worker/repository"
)

type IAccuralClient interface {
	GetResult(orderNumber string) (*model.AccuralResponse, error)
	Register(orderNumber string) error
}

type IRepo interface {
	GetNewOrders(ctx context.Context) ([]string, error)
	GetNewUnregisteredOrders(ctx context.Context) ([]string, error)
	UpdateStatus(ctx context.Context, results []model.AccuralResponse) error
	RegisterOrders(ctx context.Context, orders []string) error
}

type Worker struct {
	cfg    *Config
	client IAccuralClient
	repo   IRepo

	wg *errgroup.Group

	queueStatus   chan string
	queueRegister chan string
	statustCh     chan model.AccuralResponse
	orderCh       chan string

	doneCh chan struct{}
	stopCh chan struct{}
}

func NewWorker(ctx context.Context, cfg *Config, pool *pgxpool.Pool) *Worker {
	client := NewAccuralClient(ctx, cfg.AccrualSystemAddress)
	repo := repository.New(pool)

	return &Worker{
		cfg:    cfg,
		client: client,
		repo:   repo,

		wg: new(errgroup.Group),

		queueStatus:   make(chan string, cfg.RateLimit*2),
		queueRegister: make(chan string, cfg.RateLimit*2),
		statustCh:     make(chan model.AccuralResponse, cfg.RateLimit*2),
		orderCh:       make(chan string, cfg.RateLimit*2),

		doneCh: make(chan struct{}),
		stopCh: make(chan struct{}),
	}
}

func (w *Worker) StartWork(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.cfg.CheckNewOrderInterval) * time.Second)
	defer func() {
		ticker.Stop()
		if err := w.wg.Wait(); err != nil {
			logger.Log.Warn("", zap.Error(err))
		}

		close(w.doneCh)
	}()

	w.wg.Go(func() error {
		w.startWorks(ctx)
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				orders, err := w.repo.GetNewOrders(ctx)
				if err != nil {
					logger.Log.Warn("couldn't get orders", zap.Error(err))
					return
				}
				for _, order := range orders {
					select {
					case w.queueStatus <- order:
					case <-ctx.Done():
						return
					default:
						logger.Log.Debug("Queue is full, skipping order", zap.String("order", order))
					}
				}
			}()

			go func() {
				defer wg.Done()
				orders, err := w.repo.GetNewUnregisteredOrders(ctx)
				if err != nil {
					logger.Log.Warn("couldn't get orders", zap.Error(err))
					return
				}
				for _, order := range orders {
					select {
					case w.queueRegister <- order:
					case <-ctx.Done():
						return
					default:
						logger.Log.Debug("Queue is full, skipping order", zap.String("order", order))
					}
				}
			}()

			wg.Wait()
		}
	}
}

func (w *Worker) startWorks(ctx context.Context) {
	logger.Log.Info("Start workers", zap.Int("count", w.cfg.RateLimit))
	for i := 0; i < w.cfg.RateLimit; i++ {
		w.wg.Go(func() error {
			return w.worker(ctx)
		})

		w.wg.Go(func() error {
			return w.handleResult(ctx)
		})
	}
}

func (w *Worker) worker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-w.stopCh:
			return nil
		case batch := <-w.queueStatus:
			result, err := w.client.GetResult(batch)
			if err != nil {
				logger.Log.Warn("Failed to get accural", zap.Error(err))
				continue
			}
			if result != nil {
				w.statustCh <- *result
			}

		case batch := <-w.queueRegister:
			err := w.client.Register(batch)
			if err != nil {
				logger.Log.Warn("Failed to register order", zap.Error(err))
				continue
			}

			w.orderCh <- batch
		}
	}
}

func (w *Worker) handleResult(ctx context.Context) error {
	batchStatusResult := make([]model.AccuralResponse, 0, w.cfg.RateLimit)
	batchOrderResult := make([]string, 0, w.cfg.RateLimit)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

breakLoop:
	for {
		select {
		case <-w.stopCh:
			break breakLoop
		case <-ctx.Done():
			break breakLoop
		case result := <-w.statustCh:
			batchStatusResult = append(batchStatusResult, result)
			if len(batchStatusResult) >= w.cfg.RateLimit {
				if err := w.repo.UpdateStatus(ctx, batchStatusResult); err != nil {
					logger.Log.Warn("couldn't update orders", zap.Error(err))
				}
				batchStatusResult = batchStatusResult[:0]
			}
		case result := <-w.orderCh:
			batchOrderResult = append(batchOrderResult, result)
			if len(batchOrderResult) >= w.cfg.RateLimit {
				if err := w.repo.RegisterOrders(ctx, batchOrderResult); err != nil {
					logger.Log.Warn("couldn't update orders", zap.Error(err))
				}
				batchOrderResult = batchOrderResult[:0]
			}
		case <-ticker.C:
			if len(batchStatusResult) > 0 {
				if err := w.repo.UpdateStatus(ctx, batchStatusResult); err != nil {
					logger.Log.Warn("couldn't update orders", zap.Error(err))
				}
				batchStatusResult = batchStatusResult[:0]
			}
			if len(batchOrderResult) > 0 {
				if err := w.repo.RegisterOrders(ctx, batchOrderResult); err != nil {
					logger.Log.Warn("couldn't update orders", zap.Error(err))
				}
				batchOrderResult = batchOrderResult[:0]
			}
		}
	}

	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(batchStatusResult) > 0 {
		if err := w.repo.UpdateStatus(flushCtx, batchStatusResult); err != nil {
			logger.Log.Warn("couldn't update orders", zap.Error(err))
		}
	}
	if len(batchOrderResult) > 0 {
		if err := w.repo.RegisterOrders(flushCtx, batchOrderResult); err != nil {
			logger.Log.Warn("couldn't update orders", zap.Error(err))
		}
	}
	return nil
}

func (w *Worker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	logger.Log.Info("Worker stopped gracefully")
}

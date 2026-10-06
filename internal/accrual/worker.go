package accrual

import (
	"context"
	"time"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

// accrualWorker periodically polls the accrual system for order status updates.
type accrualWorker struct {
	client     *accrualClient
	repository Repository
	logger     *zap.Logger
	interval   time.Duration
	done       chan struct{}
}

// NewAccrualWorker creates a worker that polls the accrual system at the given interval.
func NewAccrualWorker(client *accrualClient, repository Repository, logger *zap.Logger, interval time.Duration) *accrualWorker {
	return &accrualWorker{
		client:     client,
		repository: repository,
		logger:     logger,
		interval:   interval,
		done:       make(chan struct{}),
	}
}

// Done returns a channel that is closed once ProcessNewOrders has fully stopped.
func (aw *accrualWorker) Done() <-chan struct{} {
	return aw.done
}

// ProcessNewOrders periodically fetches new orders from the repository
// and updates their accrual status until the context is cancelled.
func (aw *accrualWorker) ProcessNewOrders(ctx context.Context) error {
	defer close(aw.done)
	var (
		orders []string
		err    error
		order  accrualOrder
	)

	ticker := time.NewTicker(aw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			orders, err = aw.repository.FetchNewOrders(ctx)
			if err != nil {
				aw.logger.Error(err.Error())
				continue
			}
			for _, orderNumber := range orders {
				order, err = aw.client.GetStatus(ctx, orderNumber)
				if err != nil {
					aw.logger.Debug(err.Error(), zap.String("orderNumber", orderNumber))
					continue
				}
				if order.Status == model.StatusRegistered {
					continue
				}
				err = aw.repository.UpdateOrder(ctx, orderNumber, order.Status, order.Accrual)
				if err != nil {
					aw.logger.Error(err.Error())
					continue
				}
			}
		}
	}
}

package accrual

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// accrualWorker periodically polls the accrual system for order status updates.
type accrualWorker struct {
	client   *accrualClient
	logger   *zap.Logger
	interval time.Duration
}

// NewAccrualWorker creates a worker that polls the accrual system at the given interval.
func NewAccrualWorker(client *accrualClient, logger *zap.Logger, interval time.Duration) *accrualWorker {
	return &accrualWorker{
		client:   client,
		logger:   logger,
		interval: interval,
	}
}

// Run starts the polling loop until the context is cancelled.
func (aw *accrualWorker) Run(ctx context.Context) error {
	return nil
}

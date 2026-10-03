package accrual

import (
	"context"
)

// Repository is the storage interface consumed by accrualWorker.
type Repository interface {
	FetchNewOrders(ctx context.Context) ([]string, error)
	UpdateOrder(ctx context.Context, orderNumber string, orderStatus string, accrual float64) error
}

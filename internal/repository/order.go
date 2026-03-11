package repository

//go:generate mockgen -source=order.go -destination=mock/mock_order_repository.go -package=mock

import (
	"context"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
)

// OrderRepository — работа с заказами в БД.
type OrderRepository interface {
	GetByNumber(ctx context.Context, number string) (*models.Order, error)
	Create(ctx context.Context, userID int64, number, status string) (*models.Order, error)
	ListByUserID(ctx context.Context, userID int64) ([]*models.Order, error)
	UpdateAccrualAndStatus(ctx context.Context, number, status string, accrual *int) error
	ListNumbersPendingAccrual(ctx context.Context, statuses []string) ([]string, error)
	GetTotalAccrualsByUserID(ctx context.Context, userID int64) (int64, error)
}

package accrualclient

import (
	"context"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/usecase"

	"github.com/skiphead/go-musthave-diploma-tpl/internal/worker"
)

// Adapter адаптирует клиент API для работы с usecase
type Adapter struct {
	client *orderclient.Client
}

// NewAdapter создает новый адаптер
func NewAdapter(client *orderclient.Client) *Adapter {
	return &Adapter{client: client}
}

// GetOrderInfo возвращает информацию о заказе
func (a *Adapter) GetOrderInfo(ctx context.Context, orderNumber string) (*orderclient.OrderResponse, error) {
	// Преобразуем вызов к нужному типу
	response, err := a.client.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Ensure Adapter implements worker.OrderAPIClient
var _ worker.OrderAPIClient = (*Adapter)(nil)

// Ensure Adapter implements usecase.OrderAPI
var _ usecase.OrderAPI = (*Adapter)(nil)

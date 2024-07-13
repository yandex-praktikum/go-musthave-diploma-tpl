package accrual

import (
	"context"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/dto"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/api/orders"
	httpClient "github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/client"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
)

type OrdersAPI interface {
	SendOrder(ctx context.Context, orderDTO dto.Order) (*dto.OrderResponse, error)
}

type client struct {
	host       string
	httpClient httpClient.ClientHTTP
	logger     logging.Logger
	ordersAPI  OrdersAPI
}

func New(host string, logger logging.Logger) *client {
	httpClientInstance := httpClient.New(logger)
	ordersAPI := orders.New(httpClientInstance, host, logger)

	return &client{
		host:       host,
		httpClient: httpClientInstance,
		logger:     logger,
		ordersAPI:  ordersAPI,
	}
}

func (c *client) SendOrder(ctx context.Context, orderDTO dto.Order) (*dto.OrderResponse, error) {
	return c.ordersAPI.SendOrder(ctx, orderDTO)
}

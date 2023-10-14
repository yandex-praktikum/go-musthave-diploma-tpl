package orders

import (
	"context"
	"github.com/s-lyash/go-musthave-diploma-tpl/config"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/orders"
)

type Service struct {
	orders      orders.Repository
	accrual     chan string
	orderStatus chan models.LoyaltySystem
	conf        *config.Config
}

type Order interface {
	CheckOrder(string2 string)
	GetOrders()
	CreateOrder(string, []byte) int
	AccrualOrderStatus(context.Context) (context.Context, func())
	ChangeOrderStatus(context.Context) (context.Context, func())
}

func NewService(orders orders.Repository, conf *config.Config) Order {
	return &Service{
		orders:      orders,
		accrual:     make(chan string, 10),
		orderStatus: make(chan models.LoyaltySystem, 10),
		conf:        conf,
	}
}

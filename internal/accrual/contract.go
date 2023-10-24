package accrual

import (
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type orders interface {
	CreateOrder(userID int, num, status string) (int, time.Time, error)
	GetOrders(userID int) ([]models.Order, error)
	GetOrdersWithStatus() ([]models.OrderResponse, error)
	ChangeStatusAndSum(sum float64, status, num string) error
}

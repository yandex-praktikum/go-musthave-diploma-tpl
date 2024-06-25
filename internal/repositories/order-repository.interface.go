package repositories

import (
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
)

type OrderRepositoryInterface interface {
	UpdateOrderByAccrualOrder(accrualOrder *models.AccrualOrderResponse) error
	GetOrdersForProcess() ([]*entities.Order, error)
}

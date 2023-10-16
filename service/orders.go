package service

import (
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type OrdersService struct {
	repo repository.Orders
}

func (r *OrdersService) GetOrdersWithStatus() ([]models.OrderResponse, error) {
	return r.repo.GetOrdersWithStatus()
}

func (r *OrdersService) ChangeStatusAndSum(sum float64, status, num string) error {
	return r.repo.ChangeStatusAndSum(sum, status, num)
}

func (r *OrdersService) CreateOrder(userID int, num, status string) (int, time.Time, error) {
	return r.repo.CreateOrder(userID, num, status)
}

func (r *OrdersService) GetOrders(userID int) ([]models.Order, error) {
	return r.repo.GetOrders(userID)
}

func NewOrdersStorage(repo repository.Orders) *OrdersService {
	return &OrdersService{repo: repo}
}

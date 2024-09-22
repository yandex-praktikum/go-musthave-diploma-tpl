package service

import (
	"github.com/sub3er0/gophermart/internal/repository"
)

type OrderService struct {
	OrderRepository *repository.OrderRepository
}

func (or *OrderService) IsOrderExist(orderNumber string, userID int) (int, error) {
	return or.OrderRepository.IsOrderExist(orderNumber, userID)
}

func (or *OrderService) SaveOrder(orderNumber string, userID int, accrual int) error {
	return or.OrderRepository.SaveOrder(orderNumber, userID, accrual)
}

func (or *OrderService) GetUserOrders(userID int) ([]repository.OrderData, error) {
	return or.OrderRepository.GetUserOrders(userID)
}

func (or *OrderService) GetUserBalance(userID int) (repository.UserBalance, error) {
	return or.OrderRepository.GetUserBalance(userID)
}

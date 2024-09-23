package service

import (
	"gophermart/internal/repository"
)

type OrderService struct {
	OrderRepository *repository.OrderRepository
}

func (or *OrderService) IsOrderExist(orderNumber string, userID int) (int, error) {
	return or.OrderRepository.IsOrderExist(orderNumber, userID)
}

func (or *OrderService) SaveOrder(orderNumber string, userID int) error {
	return or.OrderRepository.SaveOrder(orderNumber, userID)
}

func (or *OrderService) UpdateOrder(orderNumber string, accrual float32, status string) error {
	return or.OrderRepository.UpdateOrder(orderNumber, accrual, status)
}

func (or *OrderService) GetUserOrders(userID int) ([]repository.OrderData, error) {
	return or.OrderRepository.GetUserOrders(userID)
}

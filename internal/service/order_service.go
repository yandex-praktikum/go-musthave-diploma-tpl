package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gophermart/internal/repository"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
}

func NewOrderService(orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

type CreateOrderResult struct {
	Created bool
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, number string) (*CreateOrderResult, error) {
	if number == "" {
		return nil, ErrInvalidInput
	}

	if !IsValidOrderNumber(number) {
		return nil, ErrInvalidOrderNumber
	}

	existingUserID, err := s.orderRepo.GetByNumber(ctx, number)
	if err == nil {
		if existingUserID == userID {
			return &CreateOrderResult{Created: false}, nil // already exists for this user
		}
		return nil, ErrConflict
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("get order by number: %w", err)
	}

	if err := s.orderRepo.Create(ctx, userID, number, "NEW", time.Now()); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return &CreateOrderResult{Created: true}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, userID int64) ([]repository.Order, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}
	return orders, nil
}

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

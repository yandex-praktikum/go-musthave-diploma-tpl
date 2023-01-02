package storage

import (
	"context"
	"fmt"

	"github.com/brisk84/gofemart/domain"
)

func (s *storage) CreateOrder(_ context.Context, order domain.Order) (int64, error) {
	s.orderMtx.Lock()
	defer s.orderMtx.Unlock()

	id := s.currentOrderID
	s.currentOrderID++
	order.ID = id
	if _, ok := s.orders[id]; ok {
		return 0, fmt.Errorf("order with id %d already exists", id)
	}
	s.orders[id] = order
	return id, nil
}
func (s *storage) GetOrder(_ context.Context, orderID int64) (domain.Order, error) {
	s.orderMtx.RLock()
	defer s.orderMtx.RUnlock()

	order, ok := s.orders[orderID]
	if !ok {
		return domain.Order{}, fmt.Errorf("order with id %d is not exists", orderID)
	}
	return order, nil
}

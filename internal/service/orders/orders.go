package orders

import (
	"context"
	"errors"
	"fmt"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/domain"
	"gmart/internal/dto"
)

// ============ Class ============

var (
	ErrInvalidOrderFormat = errors.New("invalid order format")
	ErrOrderListIsEmpty   = errors.New("order list is empty")
)

// Orders класс манипулирования заказами
type Orders struct {
	ordersRepo *OrdersRepo
}

// NewOrders создает новый объект манипулирования заказами
func NewOrders(cfg config.Config, pg pgc.PgInstance, m OrdersMetrics) (*Orders, error) {
	return &Orders{
		ordersRepo: NewOrdersRepo(pg, m),
	}, nil
}

// ============ UseCase ============

// Upload регистрирует новый заказ пользователя
func (o *Orders) Upload(ctx context.Context, userID domain.UserID, orderNumber domain.OrderNumber) error {

	// Загрузка заказа в БД
	err := o.ordersRepo.Upload(ctx, userID, orderNumber)
	if err != nil {
		return fmt.Errorf("order upload fail: %w", err)
	}

	return nil
}

// List возвращает список загруженных пользователем заказов
func (o *Orders) List(ctx context.Context, userID domain.UserID) (ordersList []dto.OrderItem, err error) {

	ordersList, err = o.ordersRepo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("order list fail: %w", err)
	}

	// Для случая пустого списка вернем проинициализированный слайс, чтобы он смаршалился в '[]'
	if len(ordersList) == 0 {
		return nil, ErrOrderListIsEmpty
	}

	return ordersList, nil
}

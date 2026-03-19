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

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=orders_mock.go  -package=orders

// OrdersRepoIFace описывает методы для работы с заказами.
// Используется для генерации моков и подмены реализации в тестах сервиса.
type OrdersRepoIFace interface {
	// Upload регистрирует новый номер заказа.
	// Возвращает ErrOrderConflict или ErrOrderAlreadyUploaded при коллизиях.
	Upload(ctx context.Context, userID domain.UserID, orderNumber domain.OrderNumber) error

	// List возвращает историю заказов конкретного пользователя.
	List(ctx context.Context, userID domain.UserID) ([]dto.OrderItem, error)

	// AcquireNextOrder извлекает следующий заказ для обработки (SKIP LOCKED).
	// Используется воркерами для синхронизации с внешним API.
	AcquireNextOrder(ctx context.Context) (domain.OrderNumber, domain.OrderStatus, error)

	// UpdateOrderStatus фиксирует результат обработки заказа (статус и баллы).
	UpdateOrderStatus(ctx context.Context, orderNumber domain.OrderNumber, orderStatus domain.OrderStatus, accrual domain.Amount) error
}

// ============ Class ============

var (
	ErrInvalidOrderFormat = errors.New("invalid order format")
	ErrOrderListIsEmpty   = errors.New("order list is empty")
)

// Orders класс манипулирования заказами
type Orders struct {
	ordersRepo OrdersRepoIFace
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

//go:generate mockgen -source=order_service.go -destination=mocks/mock_order_service.go -package=mocks
package services

import (
	"context"

	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"
)

type (
	/*
		Объект службы для работы с заказами
	*/
	OrderActions struct {
		db repository.SQLStorageInterface
	}

	/*
		Интерфейс оОъекта службы заказов
	*/
	OrderActionsInterface interface {
		Store(ctx context.Context, orderID string) error
	}
)

/*
Конструктор службы для операций с заказами
*/
func NewOrderActions(db repository.SQLStorageInterface) OrderActionsInterface {
	return &OrderActions{
		db: db,
	}
}

/*
Сохраняет закза в БД и отправляет его на обработку в Accrual
*/
func (s *OrderActions) Store(ctx context.Context, idStr string) error {
	claims := middleware.GetClaims(ctx)
	order, err := models.NewOrder(idStr, claims.UserID, models.OrderNewStatus)
	if err != nil {
		return err
	}

	if err := order.Validate(); err != nil {
		return err
	}

	// return s.db.AddOrder(order)

	return nil
}

/*
func (s *OrderActions) Add2Queue(order *models.Order) error {
	return nil
}

func (s *OrderActions) GetFromQueue() *models.Order {
	return nil
}

func (s *OrderActions) RestoreQueue(ctx context.Context) error {
	// get orders from DB where status in (new, processing)
	return nil
}

func (s *OrderActions) StartAccrualWorkerPool(ctx context.Context, n int) {
}
*/

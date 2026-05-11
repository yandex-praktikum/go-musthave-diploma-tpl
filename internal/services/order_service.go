//go:generate mockgen -source=order_service.go -destination=mocks/mock_order_service.go -package=mocks
package services

import (
	"context"
	"errors"
	"time"

	"github.com/KalessinD/gophermart/internal/clients"
	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/processors"
	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/KalessinD/gophermart/internal/services/alg"
)

type (
	/*
		Объект службы для работы с заказами
	*/
	OrderService struct {
		db            repository.SQLStorageInterface
		accrualClient clients.AccrualClienttInterface
		reportOrderCh chan *processors.Task
	}

	/*
		Интерфейс оОъекта службы заказов
	*/
	OrderServiceInterface interface {
		Store(ctx context.Context, orderID string) error
		List(ctx context.Context) (models.OrdersList, error)
		// GetOrderAccrual(ctx context.Context, order *models.Order) (*models.AccrualResponse, error)
		ProcessAccrualTask(ctx context.Context, pauseCh chan time.Duration, task *processors.Task) error
	}
)

/*
Конструктор службы для операций с заказами
*/
func NewOrderService(db repository.SQLStorageInterface, client clients.AccrualClienttInterface, outCh chan *processors.Task) OrderServiceInterface {
	return &OrderService{
		db:            db,
		accrualClient: client,
		reportOrderCh: outCh,
	}
}

/*
Сохраняет заказ в БД и отправляет его на обработку в Accrual
*/
func (s *OrderService) Store(ctx context.Context, idStr string) error {
	if !alg.IsValidLuhn(idStr) {
		return models.ErrOrderWrongFormat
	}

	claims := middleware.GetClaims(ctx)
	order, err := models.NewOrder(idStr, claims.UserID, models.OrderNewStatus)
	if err != nil {
		return err
	}

	if err := order.Validate(); err != nil {
		return err
	}

	err = s.db.AddOrder(ctx, order)
	if err != nil {
		if !errors.Is(err, models.ErrOrderExists) {
			return err
		}

		_, err = s.db.GetOrder(ctx, order.ID, order.UserID)

		switch {
		case errors.Is(err, models.ErrOrderNotFound):
			return models.ErrOrderBelongsToOtherUser
		case err == nil:
			return models.ErrOrderExists
		default:
			return err
		}
	}

	s.sendOrderToProcessor(order)

	return nil
}

func (s *OrderService) List(ctx context.Context) (models.OrdersList, error) {
	claims := middleware.GetClaims(ctx)
	orders, err := s.db.ListOrders(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) ProcessAccrualTask(ctx context.Context, pauseCh chan time.Duration, task *processors.Task) error {
	order := models.Order(*task)
	resp, err := s.accrualClient.GetOrderAccrual(ctx, task.ID)
	if err != nil {
		if errors.Is(err, clients.ErrServiceIsBusy) {
			castedErr, ok := err.(clients.AccrualError)
			if ok {
				delay := castedErr.GetDelay()
				select {
				case pauseCh <- delay:
					// отправили сигнал о паузе всем воркерам
				default:
					// если заблокировали, значит отправил сигнал сосед, спокойно идем дальше по своим делам
				}

				// вернём задачу в очередь
				s.sendOrderToProcessor(&order)
			}
		}
		return err
	}

	if resp.Status == models.OrderInProcessStatus && order.Status == resp.Status {
		// заказ в обработке, при этом мы знаем этот статус - нечего писать в БД его ещё раз
		s.sendOrderToProcessor(&order)
		return nil
	}

	targetStatus := resp.Status

	if resp.Status == models.OrderRegisteredStatus {
		// заказ зарегистрирован и ещё не начинал обрабатываться, у нас это models.OrderNewStatus
		targetStatus = models.OrderNewStatus
	}

	err = order.SetStatus(targetStatus)
	if err != nil {
		return err
	}

	if len(resp.Accrual) > 0 {
		err = order.SetAccrual(resp.Accrual.String())
		if err != nil {
			return err
		}
	}

	err = s.db.UpdateOrder(ctx, &order)
	if err != nil {
		return err
	}

	if order.Status == models.OrderInProcessStatus {
		// заказ не обработан ещё => вернём в очередь
		s.sendOrderToProcessor(&order)
	}

	return nil
}

func (s *OrderService) sendOrderToProcessor(order *models.Order) {
	task := processors.Task(*order)

	select {
	case s.reportOrderCh <- &task:
		// успех успешный
	default:
		// канал переполнен, в базе заказ уже есть, так что как-нибудь потом из БД прочтем и дообработаем
	}
}

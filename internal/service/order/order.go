package order

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/dto"
	orders2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/api/orders"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http/rest/user/orders/converter"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"golang.org/x/sync/errgroup"
	"time"
)

type Accrual interface {
	SendOrder(ctx context.Context, orderDTO dto.Order) (*dto.OrderResponse, error)
}

type Storage interface {
	Create(ctx context.Context, userID int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
	GetByOrderNumber(ctx context.Context, orderNumber *entity.OrderNumber) (*entity.OrderDB, error)
	GetOrdersForProcessing(ctx context.Context) ([]*entity.OrderDB, error)
	Update(ctx context.Context, orderDB *entity.OrderDB) error
	GetOrdersByUserID(ctx context.Context, userID int) ([]*entity.OrderDB, error)
}

type UserStorage interface {
	IncrementBalance(ctx context.Context, userID int, incValue float64) (*entity.UserDB, error)
}

type userService struct {
	accrual     Accrual
	logger      logging.Logger
	storage     Storage
	userStorage UserStorage
	cfg         *config.Config
}

var (
	ErrOrderNumberAlreadyUploadByCurrentUser = errors.New("номер заказа уже был загружен этим пользователем")
	ErrOrderNumberAlreadyUploadByOtherUser   = errors.New("номер заказа уже был загружен другим пользователем")
)

// AccrualServiceCronTime каждые 5сек опрашивать сервис на статус заказа
const AccrualServiceCronTime = time.Second * 5

func NewOrderService(accrual Accrual, logger logging.Logger, storage Storage, cfg *config.Config, userStorage UserStorage) *userService {
	return &userService{
		accrual:     accrual,
		logger:      logger,
		storage:     storage,
		cfg:         cfg,
		userStorage: userStorage,
	}
}

func (u userService) Create(ctx context.Context, userID int, orderNumber *entity.OrderNumber) (*entity.OrderDB, error) {
	orderDB, err := u.storage.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if orderDB != nil {
		if orderDB.UserID == userID {
			return nil, ErrOrderNumberAlreadyUploadByCurrentUser
		} else {
			return nil, ErrOrderNumberAlreadyUploadByOtherUser
		}
	}

	orderDB, err = u.storage.Create(ctx, userID, orderNumber)
	if err != nil {
		return nil, err
	}

	return orderDB, nil
}

func (u userService) StartProcessingOrders() {
	u.logger.Info("[StartProcessingOrders]: Старт")
	isCustomTimeSet := false
	ctx := context.Background()
	ticker := time.NewTicker(AccrualServiceCronTime)

	for range ticker.C {
		err := u.processingOrders(ctx)
		var tooManyRequestsError *orders2.TooManyRequestsError
		// при ошибке tooManyRequestsError - устанавливает новое время в тикере
		if errors.As(err, &tooManyRequestsError) {
			ticker.Reset(time.Second * time.Duration(tooManyRequestsError.RetryAfter))
			isCustomTimeSet = true
			// Если ошибка не из-за большого кол-ва запросов, то сбрасывает до обычного времени
			// если было изменено время опроса
		} else if isCustomTimeSet {
			isCustomTimeSet = false
			ticker.Reset(AccrualServiceCronTime)
		}

	}
}

// processingOrders Условно крон, который забирает все заказы, которые должна быть прогоняты через сервис accurate
func (u userService) processingOrders(ctx context.Context) error {
	u.logger.Info("[processingOrders]: Получение заказов для обработки")
	// создаём переменную errgroup
	g := new(errgroup.Group)
	orders, err := u.storage.GetOrdersForProcessing(ctx)
	if err != nil {
		u.logger.Error(err)
		return err
	}

	u.logger.Infof("[processingOrders]: Кол-во заказов: %v", len(orders))

	for _, order := range orders {
		order := order
		g.Go(func() error {
			return u.processingOrder(order)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (u userService) processingOrder(order *entity.OrderDB) error {
	u.logger.Infof("[processingOrder]: обработка заказа: %+v", order)
	ctx := context.Background()

	orderDTO := dto.Order{Number: order.Number}

	u.logger.Infof("[processingOrder]: получение данных из accrual по заказу %v", orderDTO)
	orderResponse, err := u.accrual.SendOrder(ctx, orderDTO)
	if err != nil {
		u.logger.Errorf("[processingOrder]: Ошибка получения статуса ордера из сервиса accrual: %v", err)
		return err
	}

	u.logger.Infof("[processingOrder]: Данные из accrual успешно получены %v", orderResponse)

	if orderResponse.Status != order.Status {
		newOrderDB := &entity.OrderDB{
			ID:     order.ID,
			Number: order.Number,
			Status: orderResponse.Status,
			Accrual: sql.NullFloat64{
				Float64: float64(orderResponse.Accrual),
				Valid:   true,
			},
			UploadedAt: order.UploadedAt,
			UserID:     order.UserID,
		}

		u.logger.Infof("[processingOrder]: Обновление заказа в репозитории %v", *newOrderDB)

		err := u.storage.Update(ctx, newOrderDB)
		if err != nil {
			u.logger.Errorf("[processingOrder]: Ошибка при обновлении %+v заказа в репозитории: %v", newOrderDB, err)
			return err
		}
		_, err = u.userStorage.IncrementBalance(ctx, newOrderDB.UserID, newOrderDB.Accrual.Float64)
		if err != nil {
			u.logger.Errorf("[processingOrder]: Ошибка при увеличении баланса пользователя %v", newOrderDB.UserID)
			return err
		}
	} else {
		u.logger.Infof("[processingOrder]: заказ не обновлен, статус остался таким же %v == %v", orderResponse.Status, order.Status)
	}

	u.logger.Infof("[processingOrder]: Заказ обработан: %+v", order)

	return nil
}

func (u userService) GetOrdersStatusJSONs(ctx context.Context, userID int) ([]*entity.OrderStatusJSON, error) {
	orders, err := u.storage.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ordersStatusJSONs := converter.GetOrdersStatusJSONsByOrderDBs(orders)

	return ordersStatusJSONs, nil
}

package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewOrder(storage OrderStorage) *order {
	return &order{
		storage: storage,
	}
}

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . OrderStorage
type OrderStorage interface {
	Upload(data *domain.OrderData) error
	Orders(userID int) ([]domain.OrderData, error)
	Update(number domain.OrderNumber, status domain.OrderStatus, accrual *float64) error
	ForProcessing(statuses []domain.OrderStatus) ([]domain.OrderData, error)
}

type order struct {
	storage OrderStorage
}

// Загрузка пользователем номера заказа для расчёта
// Возвращает:
//  - nil -  новый номер заказа принят в обработку
//  - domain.ErrOrderNumberAlreadyProcessed -  номер заказа уже был загружен этим пользователем;
//  - domain.ErrWrongOrderNumber - неверный формат номера заказа
//  - domain.ErrUserIsNotAuthorized - пользователь не авторизован
//  - domain.ErrDublicateOrderNumber - номер заказа уже был загружен другим пользователем
//  - domain.ErrServerInternal - внутренняя ошибка

func (ord *order) New(ctx context.Context, number domain.OrderNumber) error {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("order.New", "err", err.Error())
		return domain.ErrUserIsNotAuthorized
	}

	if !domain.CheckLuhn(number) {
		logger.Errorw("order.New", "err", fmt.Sprintf("%v: %v wrong value", domain.ErrWrongOrderNumber.Error(), number))
		return fmt.Errorf("%w: %v wrong value", domain.ErrWrongOrderNumber, number)
	}

	orderData := &domain.OrderData{
		UserID:     userID,
		Number:     number,
		Status:     domain.OrderStratusNew,
		UploadedAt: domain.RFC3339Time(time.Now()),
	}

	err = ord.storage.Upload(orderData)
	if err != nil {
		logger.Infow("order.New", "err", err.Error())
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	return nil
}

// Получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// Возвращает ошибки:
//   - domain.ErrNotFound
//   - domain.ErrUserIsNotAuthorized
//   - domain.ErrServerInternal
func (ord *order) All(ctx context.Context) ([]domain.OrderData, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("order.All", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	data, err := ord.storage.Orders(userID)
	if err != nil {
		logger.Errorw("order.All", "err", err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if data == nil {
		logger.Errorw("order.All", "err", domain.ErrNotFound.Error())
		return nil, domain.ErrNotFound
	}

	return data, nil
}

package app

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewOrders(storage OrderStorage) *orders {
	return &orders{
		storage: storage,
	}
}

type OrderStorage interface {
	Upload(userId int, data *domain.OrderData) error
	Orders(userId int) ([]domain.OrderData, error)
}

type orders struct {
	storage OrderStorage
}

// Загрузка пользователем номера заказа для расчёта
// Возвращает:
//  - nil -  новый номер заказа принят в обработку
//  - domain.ErrDublicateData -  номер заказа уже был загружен этим пользователем;
//  - domain.ErrWrongOrderNumber - неверный формат номера заказа
//  - domain.ErrUserIsNotAuthorized - пользователь не авторизован
//  - domain.ErrDublicateData - номер заказа уже был загружен другим пользователем
//  - domain.ErrServerInternal - внутренняя ошибка

func (ord *orders) Upload(ctx context.Context, number domain.OrderNumber) error {
	/*logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userId, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("orders.Upload", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	if !CheckLuhn(val) {
		logger.Errorw("balance.Upload", "err", fmt.Sprintf("%w: %v wrong value", ErrDataFormat, val))
		return fmt.Errorf("%w: %v wrong value", ErrDataFormat, val)
	}

	data := &domain.OrderData{
		Number:     number,
		Status:     domain.OrderStratusNew,
		UploadedAt: time.Now(),
	}

	err := ord.storage.Upload(userId, data) */

	return nil
}

// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
func (ord *orders) Orders(ctx context.Context) ([]domain.OrderData, error) {
	/*	logger, err := domain.GetLogger(ctx)
		if err != nil {
			log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
			return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
		}

		userId, err := domain.GetUserID(ctx)
		if err != nil {
			logger.Errorw("orders.Orders", "err", err.Error())
			return nil, domain.ErrUserIsNotAuthorized
		}
	*/
	return nil, nil
}

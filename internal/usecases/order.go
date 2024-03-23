package usecases

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/storage"
)

var (
	ErrLuhnCheck           = errors.New("invalid order format (failed Luhn alg)")
	ErrCreateExistingOrder = errors.New("order has been created already")
	ErrUniqueOrder         = errors.New("order was created by another user")
)

type OrderUseCases struct {
	strg   storage.OrderStorage
	logger logrus.FieldLogger
}

func NewOrderUseCases(strg storage.OrderStorage, logger logrus.FieldLogger) *OrderUseCases {
	return &OrderUseCases{
		strg:   strg,
		logger: logger,
	}
}

func (u *OrderUseCases) Create(ctx context.Context, userID uuid.UUID, strOrder string) (*entities.Order, error) {
	if err := checkLuhnAlg(strOrder); err != nil {
		u.logger.WithError(err).Info("checkLuhnAlg error")
		return nil, err
	}

	orderID, err := strconv.ParseUint(strOrder, 0, 64)
	if err != nil {
		u.logger.WithError(err).Info("Parse order number error")
		return nil, err
	}

	order := entities.Order{
		ID:     orderID,
		UserID: userID,
		Status: entities.StatusNew,
	}

	usrDB, err := u.strg.GetUserIDByOrder(ctx, order.ID)
	if usrDB != uuid.Nil {
		if userID == usrDB {
			return nil, ErrCreateExistingOrder
		} else {
			return nil, ErrUniqueOrder
		}
	}

	orderDB, err := u.strg.CreateOrder(ctx, &order)
	if err != nil {
		u.logger.WithError(err).Info("CreateOrder error")
		return nil, err
	}

	return orderDB, nil
}

func checkLuhnAlg(orderNum string) error {
	var sum int
	nDigits := len(orderNum)
	odd := nDigits % 2

	for i, char := range orderNum {
		digit, _ := strconv.Atoi(string(char))
		if i%2 == odd {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	if sum%10 != 0 {
		return ErrLuhnCheck
	}
	return nil
}

func (u *OrderUseCases) GetOrders(ctx context.Context, userID uuid.UUID) ([]entities.Order, error) {
	orders, err := u.strg.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (u *OrderUseCases) GetAccruals(ctx context.Context, userID uuid.UUID) (int64, error) {
	ordersDB, err := u.strg.GetAccruals(ctx, userID)
	if err != nil {
		return 0, err
	}

	return ordersDB, nil
}

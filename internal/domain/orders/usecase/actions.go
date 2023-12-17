package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/benderr/gophermart/internal/domain/orders"
	"github.com/benderr/gophermart/internal/logger"
)

type orderUsecase struct {
	orderRepo   OrderRepo
	balanceRepo BalanceRepo
	transactor  Transactor
	logger      logger.Logger
}

func New(op OrderRepo, br BalanceRepo, t Transactor, l logger.Logger) *orderUsecase {
	return &orderUsecase{
		orderRepo:   op,
		balanceRepo: br,
		transactor:  t,
		logger:      l}
}

func (o *orderUsecase) ChangeStatus(ctx context.Context, number string, status orders.Status, accrual float64) error {
	return o.transactor.Within(ctx, func(ctx context.Context, tx *sql.Tx) error {
		order, err := o.orderRepo.GetByNumber(ctx, number)
		if err != nil {
			return err
		}
		if order == nil {
			return nil
		}

		err = o.orderRepo.UpdateStatus(ctx, tx, order.Number, status)

		if err != nil {
			return err
		}

		if accrual > 0 {
			err = o.orderRepo.UpdateAccrual(ctx, tx, order.Number, accrual)
			if err != nil {
				return err
			}
		}

		if status == orders.PROCESSED && accrual > 0 {
			err = o.balanceRepo.Add(ctx, tx, order.UserId, accrual)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (o *orderUsecase) Create(ctx context.Context, userid string, number string, status orders.Status) (*orders.Order, error) {
	exist, err := o.orderRepo.GetByNumber(ctx, number)

	if err != nil {
		if errors.Is(err, orders.ErrNotFound) {
			ord, err2 := o.orderRepo.Create(ctx, userid, number, status)
			if err2 != nil {
				return nil, err2
			}
			return ord, nil
		}

		return nil, err
	}

	if exist == nil {
		return nil, orders.ErrUnexpectedFlow
	}

	if exist.UserId == userid {
		return exist, orders.ErrExistForUser
	} else {
		return exist, orders.ErrForeignForUser
	}

}

func (o *orderUsecase) GetOrdersByUser(ctx context.Context, userid string) ([]orders.Order, error) {
	return o.orderRepo.GetOrdersByUser(ctx, userid)
}

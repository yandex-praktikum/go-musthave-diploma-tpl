package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/benderr/gophermart/internal/domain/balance"
	"github.com/benderr/gophermart/internal/logger"
)

type balanceUsecase struct {
	balanceRepo   BalanceRepo
	withdrawsRepo WithdrawsRepo
	transactor    Transactor
	logger        logger.Logger
}

func New(br BalanceRepo, wr WithdrawsRepo, t Transactor, l logger.Logger) *balanceUsecase {
	return &balanceUsecase{
		balanceRepo:   br,
		withdrawsRepo: wr,
		transactor:    t,
		logger:        l}
}

func (b *balanceUsecase) Withdraw(ctx context.Context, userid string, number string, withdraw float64) error {
	return b.transactor.Within(ctx, func(ctx context.Context, tx *sql.Tx) error {

		bal, err := b.balanceRepo.GetBalanceByUser(ctx, tx, userid)

		if err != nil {
			if errors.Is(err, balance.ErrNotFound) {
				return balance.ErrInsufficientFunds
			}
			return err
		}

		if bal == nil {
			return balance.ErrUnexpectedFlow
		}

		if bal.Current < withdraw {
			return balance.ErrInsufficientFunds
		}

		err = b.withdrawsRepo.Create(ctx, tx, userid, number, withdraw)

		if err != nil {
			return err
		}

		return b.balanceRepo.Withdraw(ctx, tx, userid, withdraw)
	})

}

func (b *balanceUsecase) GetBalanceByUser(ctx context.Context, userid string) (*balance.Balance, error) {
	var resBal *balance.Balance
	err := b.transactor.Within(ctx, func(ctx context.Context, tx *sql.Tx) error {
		bal, err := b.balanceRepo.GetBalanceByUser(ctx, tx, userid)
		if err != nil {
			if errors.Is(err, balance.ErrNotFound) {
				resBal = &balance.Balance{Current: 0, Withdrawn: 0}
				return nil
			}
			return err
		}
		resBal = bal
		return nil
	})

	return resBal, err
}

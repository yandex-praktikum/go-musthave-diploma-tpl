package app

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewBalance(balanceStorage BalanceStorage) *balance {
	return &balance{
		balanceStorage: balanceStorage,
	}
}

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . BalanceStorage
type BalanceStorage interface {
	Balance(UserID int) (*domain.UserBalance, error)
	Withdraw(newBalance *domain.UserBalance, withdraw *domain.WithdrawData) error
	Withdrawals(UserID int) ([]domain.WithdrawalsData, error)
}

type balance struct {
	balanceStorage BalanceStorage
}

// получение текущего баланса счёта баллов лояльности пользователя
func (b *balance) Balance(ctx context.Context) (*domain.Balance, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't get balance - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userId, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("balance.Balance", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	uBalance, err := b.getBalance(logger, userId)
	if err != nil {
		logger.Errorw("balance.Balance", "err", err.Error())
		return nil, fmt.Errorf("get balance err: %w", err)
	}

	return &uBalance.Balance, nil
}

func (b *balance) getBalance(logger domain.Logger, userId int) (*domain.UserBalance, error) {
	uBalance, err := b.balanceStorage.Balance(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUserIsNotAuthorized
		}
		if errors.Is(err, domain.ErrDBConnection) {
			return nil, domain.ErrUserIsNotAuthorized
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if uBalance == nil {
		return nil, domain.ErrUserIsNotAuthorized
	}
	return uBalance, nil
}

// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
func (b *balance) Withdraw(ctx context.Context, withdraw *domain.WithdrawData) error {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't withdraw - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userId, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("balance.Withdraw", "err", err.Error())
		return domain.ErrUserIsNotAuthorized
	}

	if withdraw == nil {
		logger.Errorw("balance.Withdraw", "err", "withdraw is nil")
		return domain.ErrServerInternal
	}

	if !domain.CheckLuhn(string(withdraw.Order)) {
		logger.Errorw("balance.Withdraw", "err", "wrong order value")
		return fmt.Errorf("%w: wrong order value", domain.ErrDataFormat)
	}

	if withdraw.Sum <= 0 {
		logger.Errorw("balance.Withdraw", "err", "wron sum value")
		return fmt.Errorf("%w: wrong sum value", domain.ErrDataFormat)
	}

	uBalance, err := b.getBalance(logger, userId)
	if err != nil {
		logger.Errorw("balance.Withdraw", "err", err.Error())
		return fmt.Errorf("withdraw err: %w", err)
	}

	newCurrentValue := uBalance.Current - withdraw.Sum
	if newCurrentValue < 0 {
		logger.Errorw("balance.Withdraw", "err", "not enough points")
		return domain.ErrNotEnoughPoints
	}

	newWithdrawn := uBalance.WithDrawn + withdraw.Sum

	newBalance := &domain.UserBalance{
		UserID: uBalance.UserID,
		Score:  uBalance.Score + 1,
		Balance: domain.Balance{
			Current:   newCurrentValue,
			WithDrawn: newWithdrawn,
		},
	}

	err = b.balanceStorage.Withdraw(newBalance, withdraw)
	if err != nil {
		logger.Errorw("balance.Withdraw", "err", err.Error())
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	return nil
}

// получение информации о выводе средств с накопительного счёта пользователем
func (b *balance) Withdrawals(ctx context.Context) ([]domain.WithdrawalsData, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: withdrawals - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userId, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("balance.Withdrawals", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	withdrawals, err := b.balanceStorage.Withdrawals(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Errorw("balance.Withdrawals", "err", err.Error())
			return nil, fmt.Errorf("%w: %v", domain.ErrUserIsNotAuthorized, err.Error())
		}
		if errors.Is(err, domain.ErrDBConnection) {
			logger.Errorw("balance.Withdrawals", "err", err.Error())
			return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
		}
		logger.Errorw("balance.Withdrawals", "err", err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if withdrawals == nil {
		logger.Errorw("balance.Withdrawals", "err", fmt.Sprintf("user by id %v not found", userId))
		return nil, fmt.Errorf("%w: user by id %v not found", domain.ErrUserIsNotAuthorized, userId)
	}

	return withdrawals, nil
}

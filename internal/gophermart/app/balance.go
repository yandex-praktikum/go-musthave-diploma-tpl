package app

import (
	"context"
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

// Получение текущего баланса счёта баллов лояльности пользователя
// Возвращает ошибки:
//   - domain.ErrServerInternal
//   - domain.ErrUserIsNotAuthorized
func (b *balance) Balance(ctx context.Context) (*domain.Balance, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't get balance - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("balance.Balance", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	uBalance, err := b.getBalance(userID)
	if err != nil {
		logger.Errorw("balance.Balance", "err", err.Error())
		return nil, fmt.Errorf("get balance err: %w", err)
	}

	return &uBalance.Balance, nil
}

func (b *balance) getBalance(userID int) (*domain.UserBalance, error) {
	uBalance, err := b.balanceStorage.Balance(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if uBalance == nil {
		return nil, fmt.Errorf("%w: balance by id %v not found", domain.ErrServerInternal, userID)
	}
	return uBalance, nil
}

// Запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
// Возвращает ошибки:
//   - domain.ErrServerInternal
//   - domain.ErrUserIsNotAuthorized
//   - domain.ErrNotEnoughPoints
//   - domain.ErrWrongOrderNumber
func (b *balance) Withdraw(ctx context.Context, withdraw *domain.WithdrawData) error {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't withdraw - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
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
		return fmt.Errorf("%w: wrong order value", domain.ErrWrongOrderNumber)
	}

	if withdraw.Sum <= 0 {
		logger.Errorw("balance.Withdraw", "err", "wron sum value")
		return fmt.Errorf("%w: wrong sum value", domain.ErrDataFormat)
	}

	uBalance, err := b.getBalance(userID)
	if err != nil {
		logger.Errorw("balance.Withdraw", "err", err.Error())
		return fmt.Errorf("withdraw err: %w", err)
	}

	newCurrentValue := uBalance.Current - withdraw.Sum
	if newCurrentValue < 0 {
		logger.Errorw("balance.Withdraw", "err", "not enough points")
		return domain.ErrNotEnoughPoints
	}

	newWithdrawn := uBalance.Balance.Withdrawn + withdraw.Sum

	newBalance := &domain.UserBalance{
		UserID: uBalance.UserID,
		Score:  uBalance.Score + 1,
		Balance: domain.Balance{
			Current:   newCurrentValue,
			Withdrawn: newWithdrawn,
		},
	}

	err = b.balanceStorage.Withdraw(newBalance, withdraw)
	if err != nil {
		logger.Errorw("balance.Withdraw", "err", err.Error())
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	return nil
}

// Получение информации о выводе средств с накопительного счёта пользователем
// Возвращает:
//   - domain.ErrServerInternal
//   - domain.ErrUserIsNotAuthorized
//   - domain.ErrNotFound
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
		logger.Errorw("balance.Withdrawals", "err", err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if withdrawals == nil {
		logger.Errorw("balance.Withdrawals", "err", fmt.Sprintf("user by id %v not found", userId))
		return nil, fmt.Errorf("%w: user by id %v not found", domain.ErrNotFound, userId)
	}

	return withdrawals, nil
}

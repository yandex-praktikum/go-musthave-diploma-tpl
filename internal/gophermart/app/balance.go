package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewBalance(conf *config.GophermartConfig, balanceStorage BalanceStorage) *balance {
	return &balance{
		balanceStorage: balanceStorage,
	}
}

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . BalanceStorage
type BalanceStorage interface {
	Balance(ctx context.Context, userID int) (*domain.UserBalance, error)
	UpdateBalanceByWithdraw(ctx context.Context, newBalance *domain.UserBalance, withdraw *domain.WithdrawData) error
	UpdateBalanceByOrder(ctx context.Context, balance *domain.UserBalance, orderData *domain.OrderData) error
	Withdrawals(ctx context.Context, userID int) ([]domain.WithdrawalData, error)
	GetByStatus(ctx context.Context, status domain.OrderStatus) ([]domain.OrderData, error)
}

type balance struct {
	balanceStorage BalanceStorage
	once           sync.Once
}

// Получение текущего баланса счёта баллов лояльности пользователя
// Возвращает ошибки:
//   - domain.ErrServerInternal
//   - domain.ErrUserIsNotAuthorized
func (b *balance) Get(ctx context.Context) (*domain.Balance, error) {
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

	uBalance, err := b.getBalance(ctx, userID)
	if err != nil {
		logger.Errorw("balance.Balance", "err", err.Error())
		return nil, fmt.Errorf("get balance err: %w", err)
	}

	return &uBalance.Balance, nil
}

func (b *balance) getBalance(ctx context.Context, userID int) (*domain.UserBalance, error) {
	uBalance, err := b.balanceStorage.Balance(ctx, userID)
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

	if !domain.CheckLuhn(withdraw.Order) {
		logger.Errorw("balance.Withdraw", "err", "wrong order value")
		return fmt.Errorf("%w: wrong order value", domain.ErrWrongOrderNumber)
	}

	if withdraw.Sum <= 0 {
		logger.Errorw("balance.Withdraw", "err", "wron sum value")
		return fmt.Errorf("%w: wrong sum value", domain.ErrDataFormat)
	}

	uBalance, err := b.getBalance(ctx, userID)
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
		BalanceId: uBalance.BalanceId,
		UserID:    uBalance.UserID,
		Release:   uBalance.Release,
		Balance: domain.Balance{
			Current:   newCurrentValue,
			Withdrawn: newWithdrawn,
		},
	}

	err = b.balanceStorage.UpdateBalanceByWithdraw(ctx, newBalance, withdraw)
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
func (b *balance) Withdrawals(ctx context.Context) ([]domain.WithdrawalData, error) {
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

	withdrawals, err := b.balanceStorage.Withdrawals(ctx, userId)
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

func (b *balance) orderDataUpdater(ctx context.Context, logger domain.Logger, orderDataChan <-chan *domain.OrderData) {
	var sleepAfterErrChan <-chan time.Time
	var orderDataChanInternal = orderDataChan
	for {
		select {
		case <-ctx.Done():
			logger.Infow("balance.orderDataUpdater", "status", "complete")
			return
		case <-sleepAfterErrChan:
			sleepAfterErrChan = nil
			orderDataChanInternal = orderDataChan
		case orderData, ok := <-orderDataChanInternal:
			if !ok {
				logger.Infow("balance.orderDataUpdater", "status", "complete")
				return
			}
			userID := orderData.UserID
			uBalance, err := b.balanceStorage.Balance(ctx, userID)
			if err != nil {
				logger.Infow("balance.orderDataUpdater", "err", err.Error())
				sleepAfterErrChan = time.After(5 * time.Second)
				orderDataChanInternal = nil
				continue
			}

			newCurrentValue := uBalance.Current + *orderData.Accrual

			newBalance := &domain.UserBalance{
				BalanceId: uBalance.BalanceId,
				UserID:    uBalance.UserID,
				Release:   uBalance.Release,
				Balance: domain.Balance{
					Current:   newCurrentValue,
					Withdrawn: uBalance.Withdrawn,
				},
			}

			orderData.Status = domain.OrderStratusProcessed
			if err = b.balanceStorage.UpdateBalanceByOrder(ctx, newBalance, orderData); err != nil {
				logger.Infow("balance.orderDataUpdater", "err", err.Error())
				sleepAfterErrChan = time.After(5 * time.Second)
				orderDataChanInternal = nil
				continue
			}
			logger.Infow("balance.orderDataUpdater", "status", "processed", "orderNum", orderData.Number)
		}
	}
}

func (b *balance) poolOrders(ctx context.Context, logger domain.Logger, orderDataChan chan<- *domain.OrderData) {
	var sleepChan <-chan time.Time
	var orderDataChanInternal chan<- *domain.OrderData = orderDataChan
	var nextOrd *domain.OrderData
Loop:
	for {
		orders, err := b.balanceStorage.GetByStatus(ctx, domain.OrderStratusProcessing)
		if err != nil {
			logger.Infow("balance.PoolOrders", "err", err.Error())
			sleepChan = time.After(5 * time.Second)
			orderDataChanInternal = nil
			clear(orders) // для защиты от мусора
		} else {
			if len(orders) == 0 {
				logger.Infow("balance.PoolOrders", "status", "no record for pool")
				sleepChan = time.After(2 * time.Second) // статусов нужных нет, ждем две секунджы
				orderDataChanInternal = nil
			} else {
				sleepChan = nil
				nextOrd = &orders[0]
			}
		}

	OrderLoop:
		for {
			select {
			case <-ctx.Done():
				logger.Infow("balance.PoolOrders", "status", "complete")
				return
			case <-sleepChan:
				sleepChan = nil
				orderDataChanInternal = orderDataChan
				continue Loop // возвращаемя в нормальное состояние
			case orderDataChanInternal <- nextOrd:
				orders = orders[1:]
				if len(orders) == 0 {
					continue Loop
				} else {
					nextOrd = &orders[1]
					continue OrderLoop
				}

			}
		}
	}
}

func (b *balance) PoolOrders(ctx context.Context) {
	b.once.Do(func() {
		go func() {
			logger, err := domain.GetLogger(ctx)
			if err != nil {
				fmt.Printf("balance.PoolOrders error - logger not found")
				return
			}
			var wg sync.WaitGroup
			defer wg.Wait()

			orderDataChan := make(chan *domain.OrderData)

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.poolOrders(ctx, logger, orderDataChan)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.orderDataUpdater(ctx, logger, orderDataChan)
			}()
		}()
	})
}

package loyalty

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type loyaltyClient interface {
	GetAccruals(ctx context.Context, orderID domain.ID) (*domain.Order, error)
}

type loyaltyStore interface {
	AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error
	GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error)
	GetOrderForAccruals(ctx context.Context) (domain.Orders, error)
	AccrualPoints(ctx context.Context, o domain.Order) error
	UpdateOrderStatus(ctx context.Context, order domain.Order) error
	GetBalance(ctx context.Context, userID domain.ID) (domain.Balance, error)
	WithdrawPoints(ctx context.Context, userID domain.ID, o domain.Operation) error
	GetWithdrawals(ctx context.Context, userID domain.ID) ([]domain.Operation, error)
}

type Implementation struct {
	loyaltyClient loyaltyClient
	store         loyaltyStore
}

func NewImplementation(ctx context.Context, loyaltyClient loyaltyClient, store loyaltyStore) *Implementation {
	i := &Implementation{
		loyaltyClient: loyaltyClient,
		store:         store,
	}

	go i.ProcessAccrual(ctx)

	return i
}

// AddOrder ...
func (i *Implementation) AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error {
	err := i.store.AddOrder(ctx, userID, order)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error) {
	return i.store.GetOrders(ctx, userID)
}

func (i *Implementation) ProcessAccrual(ctx context.Context) {
	err := i.processAccrual(ctx)
	if err != nil {
		_ = serviceerrors.AppErrorFromError(err).LogServerError(ctx)
	}
}

const workersNums = 3

// TODO подумать, как делать это из разных подов
func (i *Implementation) processAccrual(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf(ctx, "recovered from panic: %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	workChan := make(chan domain.Order, workersNums)
	for j := 0; j < workersNums; j++ {
		wg.Add(1)
		go i.AccrualPoints(ctx, workChan, &wg)
	}
	go i.startSendWork(ctx, workChan)

	wg.Wait()

	return nil
}

func (i *Implementation) startSendWork(ctx context.Context, workChan chan domain.Order) {
	ticker := time.NewTicker(3 * time.Second)
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			orders, err := i.store.GetOrderForAccruals(ctx)
			if err != nil {
				logger.Errorf(ctx, err.Error())
			}

			for _, order := range orders {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					workChan <- order
				}
			}
		}
	}

	close(workChan)
}

func (i *Implementation) AccrualPoints(ctx context.Context,
	orders chan domain.Order, wg *sync.WaitGroup) {
	defer wg.Done()
	for order := range orders {
		err := i.accrualPoints(ctx, order)
		if err != nil {
			_ = serviceerrors.AppErrorFromError(err).LogServerError(ctx)
		}
	}
}

func (i *Implementation) accrualPoints(ctx context.Context,
	order domain.Order) error {
	logger.Infof(ctx, "start accrual %v", order.ID)
	var res *domain.Order
	res, err := i.loyaltyClient.GetAccruals(ctx, order.ID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoContent):
			return nil
		}

		return err
	}
	if res == nil {
		return nil
	}

	if res.State != domain.Processed {
		logger.Infof(ctx, "update accrual state for order %d", order.ID)
		order.State = res.State
		err = i.store.UpdateOrderStatus(ctx, order)
		if err != nil {
			return err
		}
		return nil
	}

	order.State = res.State
	order.AccrualAmount = res.AccrualAmount

	err = i.store.AccrualPoints(ctx, order)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "finish accrual points for order %d", order.ID)
	return nil
}

func (i *Implementation) GetBalance(ctx context.Context, userID domain.ID) (domain.Balance, error) {
	return i.store.GetBalance(ctx, userID)
}

func (i *Implementation) WithdrawPoints(ctx context.Context, userID domain.ID, o domain.Operation) error {
	balance, err := i.store.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balance.GetCurrent().Amount.LessThan(decimal.Zero) {
		return serviceerrors.NewPaymentRequired()
	}

	return i.store.WithdrawPoints(ctx, userID, o)
}

func (i *Implementation) GetWithdrawals(ctx context.Context, userID domain.ID) ([]domain.Operation, error) {
	res, err := i.store.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, serviceerrors.NewNoContent()
	}

	return res, nil
}

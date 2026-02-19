package gophermart

import (
	"context"
	"fmt"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/accrual"
	"github.com/Raime-34/gophermart.git/internal/cfg"
	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrUserAlreadyExist  = fmt.Errorf("User already exist")
	ErrUserNotFound      = fmt.Errorf("Failed to find user with given login")
	ErrIncorrectPassword = fmt.Errorf("Incorrect password")

	ErrNotEnoughBonuses = fmt.Errorf("Not enough bonuses")
)

type Gophermart struct {
	repositories      repositoriesInt
	accrualCalculator accrualCalculator
}

func NewGophermart(ctx context.Context, connPool *pgxpool.Pool, wg *sync.WaitGroup) *Gophermart {
	calculator := accrual.NewAccrualCalculator(cfg.GetConfig().AccrualSystemUrl)
	ch := calculator.StartMonitoring(ctx, wg)

	gophermart := &Gophermart{
		repositories:      repositories.NewRepositories(connPool),
		accrualCalculator: calculator,
	}
	go gophermart.handleOrderState(ctx, ch, wg)

	return gophermart
}

func (g *Gophermart) handleOrderState(
	ctx context.Context,
	ch <-chan *dto.AccrualCalculatorDTO,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case newState, ok := <-ch:
			if !ok {
				return
			}

			reqCtx := context.WithValue(ctx, consts.UserIdKey, newState.GetUserId())

			if err := g.repositories.UpdateOrder(reqCtx, *newState); err != nil {
				logger.Error("Failed to update order", zap.Error(err))
			}
		}
	}
}

func (g *Gophermart) RegisterUser(ctx context.Context, userInfo dto.UserCredential) error {
	if _, err := g.repositories.GetUser(ctx, userInfo); err == nil {
		return ErrUserAlreadyExist
	}

	return g.repositories.RegisterUser(ctx, userInfo)
}

func (g *Gophermart) LoginUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error) {
	userData, err := g.repositories.GetUser(ctx, userInfo)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if userData.Password != userInfo.Password {
		return nil, ErrIncorrectPassword
	}

	return userData, nil
}

func (g *Gophermart) InsertOrder(ctx context.Context, orderNumber string) error {
	userId := ctx.Value(consts.UserIdKey)
	switch t := userId.(type) {
	case string:
	case uuid.UUID:
	default:
		return fmt.Errorf("Invalid userId type: %T", t)
	}

	userIdStr, _ := userId.(string)

	g.accrualCalculator.AddToMonitoring(orderNumber, userIdStr)
	return g.repositories.RegisterOrder(ctx, orderNumber)
}

func (g *Gophermart) GetUserOrders(ctx context.Context) ([]*dto.GetOrdersInfoResp, error) {
	orders, err := g.repositories.GetOrders(ctx)
	return orderInfoSliceToGetOrdersInfoResp(orders), err
}

func (g *Gophermart) GetUserBalance(ctx context.Context) (*dto.BalanceInfo, error) {
	balance := dto.BalanceInfo{}

	allBonuses, allWithdrawls, err := g.getBonusesNWithdrawls(ctx)
	if err != nil {
		return nil, err
	}

	balance.Current = float64(allBonuses - allWithdrawls)
	balance.Withdraw = allWithdrawls

	return &balance, nil
}

func (g *Gophermart) ProcessWithdraw(ctx context.Context, req dto.WithdrawRequest) error {
	allBonuses, allWithdralws, err := g.getBonusesNWithdrawls(ctx)
	if err != nil {
		return err
	}

	if allBonuses < allWithdralws {
		return ErrNotEnoughBonuses
	}

	return g.repositories.RegisterWithdraw(ctx, req)
}

func (g *Gophermart) GetWithdraws(ctx context.Context) ([]*dto.WithdrawInfo, error) {
	return g.repositories.GetWithdraws(ctx)
}

func (g *Gophermart) getBonusesNWithdrawls(ctx context.Context) (int, int, error) {
	orders, err := g.repositories.GetOrders(ctx)
	if err != nil {
		return 0, 0, err
	}
	allBonuses := 0
	for _, order := range orders {
		allBonuses += order.Accrual
	}

	withdrals, err := g.GetWithdraws(ctx)
	if err != nil {
		return 0, 0, err
	}
	allWithdralws := 0
	for _, withdrawl := range withdrals {
		allWithdralws += withdrawl.Sum
	}

	return allBonuses, allWithdralws, nil
}

func orderInfoSliceToGetOrdersInfoResp(original []*dto.OrderInfo) []*dto.GetOrdersInfoResp {
	result := []*dto.GetOrdersInfoResp{}

	for _, order := range original {
		result = append(result, order.ToGetOrdersInfoResp())
	}

	return result
}

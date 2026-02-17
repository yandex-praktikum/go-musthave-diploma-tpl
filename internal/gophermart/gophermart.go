package gophermart

import (
	"context"
	"fmt"

	"github.com/Raime-34/gophermart.git/internal/accrual"
	"github.com/Raime-34/gophermart.git/internal/cfg"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrUserAlreadyExist  = fmt.Errorf("User already exist")
	ErrUserNotFound      = fmt.Errorf("Failed to find user with given login")
	ErrIncorrectPassword = fmt.Errorf("Incorrect password")
)

type Gophermart struct {
	repositories      *repositories.Repositories
	accrualCalculator accrualCalculator
}

func NewGophermart(ctx context.Context, connPool *pgxpool.Pool) *Gophermart {
	calculator := accrual.NewAccrualCalculator(cfg.GetConfig().AccrualSystemUrl)
	ch := calculator.StartMonitoring(ctx)

	gophermart := &Gophermart{
		repositories:      repositories.NewRepositories(ctx, connPool),
		accrualCalculator: calculator,
	}
	go gophermart.handleOrderState(ch)

	return gophermart
}

func (g *Gophermart) handleOrderState(ch <-chan *dto.AccrualCalculatorDTO) {
	for newState := range ch {
		if err := g.repositories.OrderRepo.UpdateOrder(context.Background(), *newState); err != nil {
			logger.Error("Failed to update order", zap.Error(err))
		}
	}
}

func (g *Gophermart) RegisterUser(ctx context.Context, userInfo dto.UserCredential) error {
	if _, err := g.repositories.UserRepo.GetUser(ctx, userInfo); err == nil {
		return ErrUserAlreadyExist
	}

	return g.repositories.UserRepo.RegisterUser(ctx, userInfo)
}

func (g *Gophermart) LoginUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error) {
	userData, err := g.repositories.UserRepo.GetUser(ctx, userInfo)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if userData.Password != userInfo.Password {
		return nil, ErrIncorrectPassword
	}

	return userData, nil
}

func (g *Gophermart) InsertOrder(ctx context.Context, orderNumber string) error {
	return g.repositories.OrderRepo.RegisterOrder(ctx, orderNumber)
}

func (g *Gophermart) GetUserOrders(ctx context.Context) ([]*dto.GetOrdersInfoResp, error) {
	return g.repositories.OrderRepo.GetOrders(ctx)
}

// func (g *Gophermart) GetUserBalance() []dto.BalanceInfo {}

// func (g *Gophermart) ProcessWithdraw(dto.WithdrawRequest) error {}

// func (g *Gophermart) GetWithdraws() []dto.WithdrawInfo {}

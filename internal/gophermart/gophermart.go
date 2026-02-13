package gophermart

import (
	"context"
	"fmt"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserAlreadyExist  = fmt.Errorf("User already exist")
	ErrUserNotFound      = fmt.Errorf("Failed to find user with given login")
	ErrIncorrectPassword = fmt.Errorf("Incorrect password")
)

type Gophermart struct {
	repositories *repositories.Repositories
}

func NewGophermart(ctx context.Context, connPool *pgxpool.Pool) *Gophermart {
	return &Gophermart{
		repositories: repositories.NewRepositories(ctx, connPool),
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

// func (g *Gophermart) InsertOrder(int) error {}

// func (g *Gophermart) GetUserOrders() []dto.OrderInfo {}

// func (g *Gophermart) GetUserBalance() []dto.BalanceInfo {}

// func (g *Gophermart) ProcessWithdraw(dto.WithdrawRequest) error {}

// func (g *Gophermart) GetWithdraws() []dto.WithdrawInfo {}

package repositories

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/dto"
	repositoriesorders "github.com/Raime-34/gophermart.git/internal/repositories/orders"
	repositoriesusers "github.com/Raime-34/gophermart.git/internal/repositories/users"
	repositorieswithdrawals "github.com/Raime-34/gophermart.git/internal/repositories/withdrawals"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	UserRepo      repositoriesusers.UserRepo
	OrderRepo     repositoriesorders.OrderRepo
	WithdrawlRepo repositorieswithdrawals.WithdrawalsRepo
}

func NewRepositories(conn *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepo:      *repositoriesusers.NewUserRepo(conn),
		OrderRepo:     *repositoriesorders.NewOrderRepo(conn),
		WithdrawlRepo: *repositorieswithdrawals.NewWithdrawalsRepo(conn),
	}
}

func (r *Repositories) UpdateOrder(ctx context.Context, state dto.AccrualCalculatorDTO) error {
	return r.OrderRepo.UpdateOrder(ctx, state)
}

func (r *Repositories) RegisterOrder(ctx context.Context, orderNumber string) error {
	return r.OrderRepo.RegisterOrder(ctx, orderNumber)
}

func (r *Repositories) GetOrders(ctx context.Context) ([]*dto.OrderInfo, error) {
	return r.OrderRepo.GetOrders(ctx)
}

func (r *Repositories) GetUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error) {
	return r.UserRepo.GetUser(ctx, userInfo)
}

func (r *Repositories) RegisterUser(ctx context.Context, userInfo dto.UserCredential) error {
	return r.UserRepo.RegisterUser(ctx, userInfo)
}

func (r *Repositories) RegisterWithdraw(ctx context.Context, req dto.WithdrawRequest) error {
	return r.WithdrawlRepo.RegisterWithdraw(ctx, req)
}

func (r *Repositories) GetWithdraws(ctx context.Context) ([]*dto.WithdrawInfo, error) {
	return r.WithdrawlRepo.GetWithdraws(ctx)
}

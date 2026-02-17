package repositories

import (
	"context"

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

func NewRepositories(ctx context.Context, conn *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepo:      *repositoriesusers.NewUserRepo(conn),
		OrderRepo:     *repositoriesorders.NewOrderRepo(conn),
		WithdrawlRepo: *repositorieswithdrawals.NewWithdrawalsRepo(conn),
	}
}

package repositories

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/logger"
	repositoriesorders "github.com/Raime-34/gophermart.git/internal/repositories/orders"
	repositoriesusers "github.com/Raime-34/gophermart.git/internal/repositories/users"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repositories struct {
	UserRepo  repositoriesusers.UserRepo
	OrderRepo repositoriesorders.OrderRepo
}

func NewRepositories(ctx context.Context, conn *pgxpool.Pool) *Repositories {
	userConn, err := conn.Acquire(ctx)
	if err != nil {
		logger.Fatal("Error while acquiring connection from the database pool: %v", zap.Error(err))
	}

	orderConn, err := conn.Acquire(ctx)
	if err != nil {
		logger.Fatal("Error while acquiring connection from the database pool: %v", zap.Error(err))
	}

	return &Repositories{
		UserRepo:  *repositoriesusers.NewUserRepo(userConn),
		OrderRepo: *repositoriesorders.NewOrderRepo(orderConn),
	}
}

package repositories

import (
	"context"

	"github.com/Raime-34/gophermart.git/internal/logger"
	repositoriesusers "github.com/Raime-34/gophermart.git/internal/repositories/users"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repositories struct {
	UserRepo repositoriesusers.UserRepo
}

func NewRepositories(ctx context.Context, conn *pgxpool.Pool) *Repositories {
	userConn, err := conn.Acquire(ctx)
	if err != nil {
		logger.Fatal("Error while acquiring connection from the database pool: %v", zap.Error(err))
	}

	return &Repositories{
		UserRepo: *repositoriesusers.NewUserRepo(ctx, userConn),
	}
}

package repositories

import (
	"context"

	repositoriesusers "github.com/Raime-34/gophermart.git/internal/repositories/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	UserRepo repositoriesusers.UserRepo
}

func NewRepositories(ctx context.Context, conn *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepo: *repositoriesusers.NewUserRepo(ctx, conn),
	}
}

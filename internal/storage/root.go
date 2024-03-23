package storage

import (
	"context"
	"database/sql"
	"github.com/google/uuid"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/storage/postgres"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	GetUserByLogin(ctx context.Context, user *entities.User) (*entities.User, error)
}

type OrderStorage interface {
	GetUserIDByOrder(ctx context.Context, order uint64) (uuid.UUID, error)
	CreateOrder(ctx context.Context, order *entities.Order) (*entities.Order, error)
	GetOrders(ctx context.Context, userID uuid.UUID) ([]entities.Order, error)
	GetAccruals(ctx context.Context, userID uuid.UUID) (int64, error)
}

type WithdrawStorage interface {
	CreateWithdraw(ctx context.Context, withdraw *entities.Withdraw) (*entities.Withdraw, error)
	GetSumWithdrawals(ctx context.Context, userID uuid.UUID) (int64, error)
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entities.Withdraw, error)
}

type Storager struct {
	Users       UserStorage
	Orders      OrderStorage
	Withdrawals WithdrawStorage
}

func NewStorager(db *sql.DB) *Storager {
	return &Storager{
		Users:       postgres.NewAuthPSQL(db),
		Orders:      postgres.NewOrderPSQL(db),
		Withdrawals: postgres.NewWithdrawPSQL(db),
	}
}

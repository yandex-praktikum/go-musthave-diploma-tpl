package services

import (
	"context"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

type AuthStore interface {
	SelectUserByEmail(ctx context.Context, email string) (*models.User, error)
	InsertUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

type OrderStore interface {
	InsertOrder(ctx context.Context, order *models.Order) error
	UpdateOrder(ctx context.Context, order *models.Order) error
	SelectUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
	SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	SelectOrdersForProccesing(ctx context.Context) ([]*models.Order, error)
	UpdateUser(ctx context.Context, user *models.User) error
	SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type UserStore interface {
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	SelectUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type WithdrawStore interface {
	InsertWithdraw(ctx context.Context, withdraw *models.Withdraw) error
	SelectUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*models.Withdraw, error)
	SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

package repository

import (
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type Autorisation interface {
	CreateUser(user models.User) (int, error)
	GetUser(username string) (models.User, error)
}

type Orders interface {
	CreateOrder(userID int, num, status string) (int, time.Time, error)
	GetOrders(userID int) ([]models.Order, error)
	GetOrdersWithStatus() ([]models.OrderResponse, error)
	ChangeStatusAndSum(sum float64, status, num string) error
}

type Balance interface {
	GetBalance(userID int) (models.Balance, error)
	DoWithdraw(userID int, withdraw models.Withdraw) error
	GetWithdraws(userID int) ([]models.WithdrawResponse, error)
}

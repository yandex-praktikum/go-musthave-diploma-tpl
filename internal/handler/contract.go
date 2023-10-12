package handler

import (
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type Autorisation interface {
	CreateUser(user models.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Orders interface {
	CreateOrder(user_id int, num, status string) (int, time.Time, error)
	GetOrders(user_id int) ([]models.Order, error)
	GetOrdersWithStatus() ([]models.OrderResponse, error)
	ChangeStatusAndSum(sum float64, status, num string) error
}

type Balance interface {
	GetBalance(user_id int) (models.Balance, error)
	Withdraw(user_id int, withdraw models.Withdraw) error
	GetWithdraws(user_id int) ([]models.WithdrawResponse, error)
}

type Storage struct {
	Autorisation
	Orders
	Balance
}

func NewStorage(repos *repository.Repository) *Storage {
	return &Storage{
		Autorisation: NewAuthStorage(repos.Autorisation),
		Orders:       NewOrdersStorage(repos.Orders),
		Balance:      NewBalanceStorage(repos.Balance),
	}
}

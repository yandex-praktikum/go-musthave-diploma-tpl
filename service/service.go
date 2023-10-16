package service

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
	CreateOrder(userID int, num, status string) (int, time.Time, error)
	GetOrders(userID int) ([]models.Order, error)
	GetOrdersWithStatus() ([]models.OrderResponse, error)
	ChangeStatusAndSum(sum float64, status, num string) error
}

type Balance interface {
	GetBalance(userID int) (models.Balance, error)
	Withdraw(userID int, withdraw models.Withdraw) error
	GetWithdraws(userID int) ([]models.WithdrawResponse, error)
}

type Service struct {
	Autorisation
	Orders
	Balance
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Autorisation: NewAuthStorage(repos.Autorisation),
		Orders:       NewOrdersStorage(repos.Orders),
		Balance:      NewBalanceStorage(repos.Balance),
	}
}

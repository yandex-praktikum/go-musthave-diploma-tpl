package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type Autorisation interface {
	CreateUser(user models.User) (int, error)
	GetUser(username string) (models.User, error)
}

type Orders interface {
	CreateOrder(user_id int, num, status string) (int, time.Time, error)
	GetOrders(user_id int) ([]models.Order, error)
	GetOrdersWithStatus() ([]models.OrderResponse, error)
	ChangeStatusAndSum(sum float64, status, num string) error
}

type Balance interface {
	GetBalance(user_id int) (models.Balance, error)
	DoWithdraw(user_id int, withdraw models.Withdraw) error
	ExistOrder(order int) bool
	GetWithdraws(user_id int) ([]models.WithdrawResponse, error)
}

type Repository struct {
	Autorisation
	Orders
	Balance
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Autorisation: NewAuthPostgres(db),
		Orders:       NewOrdersPostgres(db),
		Balance:      NewBalancePostgres(db),
	}
}

package orders

import (
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

//go:generate mockgen -source=./service.go -destination=service_mock.go -package=orders
type Storage interface {
	GetAccrual(addressAccrual string)
	GetUserByAccessToken(order string, login string, now time.Time) error
	GetAllUserOrders(login string) ([]*models.OrdersUser, error)
	GetBalanceUser(login string) (*models.Balance, error)
	GetWithdrawals(login string) ([]*models.Withdrawals, error)
	CheckWriteOffOfFunds(login, order string, sum float32, now time.Time) error
}

type Service struct {
	db   Storage
	logs *logger.Logger
}

func NewService(db Storage, logger *logger.Logger) *Service {
	return &Service{
		db,
		logger,
	}
}

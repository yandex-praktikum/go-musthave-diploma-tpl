package orders

import (
	"database/sql"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

//go:generate mockgen -source=./service.go -destination=service_mock.go -package=orders
type Storage interface {
	Save(query string, args ...interface{}) error
	Get(query string, args ...interface{}) (*sql.Row, error)
	Gets(query string, args ...interface{}) (*sql.Rows, error)
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

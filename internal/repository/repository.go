package repository

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository/cache"

	"github.com/sirupsen/logrus"
)

type Auth interface {
	SaveUser(*models.User, uint64) error
	GetUser(*models.User) (uint64, error)
}

type Queue interface {
	AddToQueue(order string)
	TakeFirst() string
	RemoveFromQueue()
}

type Cache interface {
	AddToCache(key string, value string)
	GetFromCache(key string) (string, bool)
	RemoveFromCache(key string)
	PrintCache() string
}
type Loyalty interface {
	CreateLoyaltyAccount(uint64) error
	SaveOrder(order *models.Order, login string) error
	GetOrders(login string) ([]models.OrderDTO, error)
	UpdateOrder(*models.Order) error
	GetBalance(login string) (*models.Account, error)
	Withdraw(*models.WithdrawalDTO, string) error
	GetWithdrawls(string) ([]models.WithdrawalDTO, error)
}

type Repository struct {
	DB
	logger *logrus.Logger
	Auth
	Queue
	Cache
	Loyalty
}

func NewRepository(db DB, logger *logrus.Logger) *Repository {
	return &Repository{
		DB:      db,
		Auth:    NewAuth(db),
		Queue:   NewQueue(),
		Cache:   cache.NewCache(),
		Loyalty: NewLoyalty(db, logger),
		logger:  logger,
	}
}

package service

import (
	"Loyalty/internal/client"
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	numbergenerator "Loyalty/pkg/numberGenerator"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Repository interface {
	//auth methods
	SaveUser(*models.User, uint64) error
	GetUser(*models.User) (uint64, error)
	//user methods
	CreateLoyaltyAccount(uint64) error
	SaveOrder(order *models.Order, login string) error
	GetOrders(login string) ([]models.Order, error)
	UpdateOrder(*models.Order) error
	GetBalance(login string) (*models.Account, error)
	CheckOrder(number string, login string) (string, error)
	Withdraw(*models.Withdraw, string) error
	GetWithdrawls(string) ([]models.Withdraw, error)
	//orders queue
	AddToQueue(order string)
	TakeFirst() string
	RemoveFromQueue()
	//accrual cash
	AddToCash(key string, value string)
	GetFromCash(key string) (string, bool)
	RemoveFromCash(key string)
	PrintCash() string
}
type Client interface {
	SentOrder(order string) (*models.Accrual, error)
	AccrualMock() error
}
type Service struct {
	Repository
	Auth
	Client
	logger *logrus.Logger
}

func NewService(r *repository.Repository, c *client.AccrualClient, logger logrus.Logger) *Service {
	return &Service{
		Repository: r,
		Auth:       *NewAuth(r),
		Client:     c,
		logger:     &logger,
	}
}

func (s *Service) CreateLoyaltyAccount(user *models.User) (uint64, error) {
	//create account number
	number, err := numbergenerator.GenerateNumber(15)
	if err != nil {
		return 0, err
	}
	//save accoun in db
	if err := s.Repository.CreateLoyaltyAccount(number); err != nil {
		return 0, err
	}
	return number, nil
}

func (s *Service) UpdateOrdersQueue() {
	timeOut := time.Millisecond * time.Duration(viper.GetInt("accrual.timeout"))
	for {
		time.Sleep(1 * time.Second) //убрать
		order := s.Repository.TakeFirst()
		if order == "" {
			time.Sleep(timeOut)
			continue
		}
		accrual, err := s.Client.SentOrder(order)
		if err != nil {
			if errors.Is(err, errors.Unwrap(err)) {
				s.Repository.RemoveFromQueue()
				continue
			}

			time.Sleep(timeOut)
			s.logger.Error(err)
			continue
		}
		s.logger.Infof("Worker: %v", accrual)
		//if order status is final
		if accrual.Status == client.StatusInvalid || accrual.Status == client.StatusProcessed {
			var order models.Order
			order.Number = accrual.Order
			order.Status = accrual.Status

			if err := s.Repository.UpdateOrder(&order); err != nil {
				time.Sleep(timeOut)
				continue
			}
			s.Repository.RemoveFromCash(accrual.Order)
			s.Repository.RemoveFromQueue()
			//if order status is not final
		} else {
			status, _ := s.Repository.GetFromCash(accrual.Order)

			if accrual.Status != status {
				var order models.Order
				order.Number = accrual.Order
				order.Status = accrual.Status
				order.Accrual = accrual.Accrual
				s.Repository.UpdateOrder(&order)
				s.Repository.AddToCash(accrual.Order, accrual.Status)
			}
		}
	}
}

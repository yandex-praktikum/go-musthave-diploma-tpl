package service

import (
	"Loyalty/internal/client"
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"Loyalty/pkg/luhn"
	numbergenerator "Loyalty/pkg/numberGenerator"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const StatusNew = "NEW"

type Repository interface {
	//auth methods
	SaveUser(*models.User, uint64) error
	GetUser(*models.User) (uint64, error)
	//user methods
	CreateLoyaltyAccount(uint64) error
	SaveOrder(order *models.Order, login string) error
	GetOrders(login string) ([]models.OrderDTO, error)
	UpdateOrder(*models.Order) error
	GetBalance(login string) (*models.Account, error)
	Withdraw(*models.WithdrawalDTO, string) error
	GetWithdrawls(string) ([]models.WithdrawalDTO, error)
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

func NewService(r *repository.Repository, c *client.AccrualClient, logger *logrus.Logger) *Service {
	return &Service{
		Repository: r,
		Auth:       *NewAuth(r, logger),
		Client:     c,
		logger:     logger,
	}
}
func (s *Service) Withdraw(withdraw *models.WithdrawalDTO, login string) error {
	//validate order number
	if ok := luhn.Validate(string(withdraw.Order)); !ok {
		return ErrNotValid
	}
	//check bonuses
	accountState, err := s.Repository.GetBalance(login)
	if err != nil {
		return ErrInt
	}
	sum := uint64(withdraw.Sum * 100)
	// if not enough bonuses
	if accountState.Current < sum {
		return ErrNoMoney
	}
	//save order in db
	var order models.Order
	order.Number = string(withdraw.Order)
	order.Status = StatusNew
	order.Accrual = 0
	if err := s.Repository.SaveOrder(&order, login); err != nil {
		return ErrInt
	}
	//save withdraw in db
	if err := s.Repository.Withdraw(withdraw, login); err != nil {
		return ErrInt
	}
	return nil
}

// creating account for user
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

// updating orders queue ============================================================
func (s *Service) UpdateOrdersQueue() {
	timeOut := time.Millisecond * time.Duration(viper.GetInt("accrual.timeout"))
	for {
		//take first order from queue
		number := s.Repository.TakeFirst()
		if number == "" {
			time.Sleep(timeOut)
			continue
		}
		//sent order to accrual system
		accrual, err := s.Client.SentOrder(number)
		if err != nil {
			if errors.Is(err, errors.Unwrap(err)) {
				s.Repository.RemoveFromQueue()
				continue
			}
			time.Sleep(timeOut)
			s.logger.Error(err)
			continue
		}
		var order models.Order
		order.Number = accrual.Order
		order.Status = accrual.Status
		order.Accrual = int(accrual.Accrual * 100)

		s.logger.Infof("Worker: %v", order)
		//if order status is final
		if accrual.Status == client.StatusInvalid || accrual.Status == client.StatusProcessed {

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
				order.Accrual = int(accrual.Accrual * 100)
				s.Repository.UpdateOrder(&order)
				s.Repository.AddToCash(accrual.Order, accrual.Status)
			}
		}
	}
}

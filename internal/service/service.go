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

type Client interface {
	SentOrder(order string) (*models.Accrual, error)
}

type Service struct {
	Repository *repository.Repository
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

//making withdrawal ============================================================
func (s *Service) SaveOrder(number string, login string) error {

	var order models.Order
	order.Number = string(number)
	order.Status = models.StatusNew
	order.Accrual = 0

	//save order in db
	if err := s.Repository.Loyalty.SaveOrder(&order, login); err != nil {
		return err
	}
	//add order to queue
	s.Repository.AddToQueue(string(number))

	return nil
}

//getting orders ============================================================
func (s *Service) GetOrders(login string) ([]models.OrderDTO, error) {
	ordersList, err := s.Repository.Loyalty.GetOrders(login)
	if err != nil {
		return nil, err
	}
	return ordersList, nil
}

//getting balance ============================================================
func (s *Service) GetBalance(login string) (*models.Account, error) {
	accountState, err := s.Repository.Loyalty.GetBalance(login)
	if err != nil {
		return nil, err
	}

	return accountState, nil
}

//getting withdrawals ============================================================
func (s *Service) GetWithdrawals(login string) ([]models.WithdrawalDTO, error) {
	withdrawls, err := s.Repository.Loyalty.GetWithdrawls(login)
	if err != nil {
		return nil, err
	}

	return withdrawls, nil
}

//making withdrawal ============================================================
func (s *Service) Withdraw(withdraw *models.WithdrawalDTO, login string) error {
	//validate order number
	if ok := luhn.Validate(string(withdraw.Order)); !ok {
		return ErrNotValid
	}
	//check bonuses
	accountState, err := s.Repository.Loyalty.GetBalance(login)
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
	order.Status = models.StatusNew
	order.Accrual = 0
	if err := s.Repository.Loyalty.SaveOrder(&order, login); err != nil {
		return ErrInt
	}
	//save withdraw in db
	if err := s.Repository.Loyalty.Withdraw(withdraw, login); err != nil {
		return ErrInt
	}
	return nil
}

// creating account for user ============================================================
func (s *Service) CreateLoyaltyAccount() (uint64, error) {
	//create account number
	number, err := numbergenerator.GenerateNumber(15)
	if err != nil {
		return 0, err
	}
	//save accoun in db
	if err := s.Repository.Loyalty.CreateLoyaltyAccount(number); err != nil {
		return 0, err
	}
	return number, nil
}

// updating orders queue ============================================================
func (s *Service) UpdateOrdersQueue() {
	timeOut := time.Millisecond * time.Duration(viper.GetInt("accrual.timeout"))
	for {
		//take first order from queue
		number := s.Repository.Queue.TakeFirst()
		if number == "" {
			time.Sleep(timeOut)
			continue
		}
		//sent order to accrual system
		accrual, err := s.Client.SentOrder(number)
		if err != nil {
			if errors.Is(err, errors.Unwrap(err)) {
				s.Repository.Queue.RemoveFromQueue()
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

			if err := s.Repository.Loyalty.UpdateOrder(&order); err != nil {
				time.Sleep(timeOut)
				continue
			}
			s.Repository.Cache.RemoveFromCache(accrual.Order)
			s.Repository.Queue.RemoveFromQueue()
			//if order status is not final
		} else {
			status, _ := s.Repository.Cache.GetFromCache(accrual.Order)

			if accrual.Status != status {
				var order models.Order
				order.Number = accrual.Order
				order.Status = accrual.Status
				order.Accrual = int(accrual.Accrual * 100)
				s.Repository.Loyalty.UpdateOrder(&order)
				s.Repository.Cache.AddToCache(accrual.Order, accrual.Status)
			}
		}
	}
}

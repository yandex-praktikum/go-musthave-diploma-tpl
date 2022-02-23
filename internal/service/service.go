package service

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	numbergenerator "Loyalty/pkg/numberGenerator"
)

type Repository interface {
	//auth methods
	SaveUser(*models.User, uint64) error
	GetUser(*models.User) (uint64, error)
	//user methods
	CreateLoyaltyAccount(uint64) error
	SaveOrder(order *models.Order, login string) error
	GetOrders(login string) ([]models.Order, error)
	GetBalance(login string) (*models.Account, error)
	CheckOrder(number string, login string) (string, error)
	Withdraw(*models.Withdraw, string) error
	GetWithdrawls(string) ([]models.Withdraw, error)
}

type Service struct {
	Repository
	Auth
}

func NewService(r *repository.Repository) *Service {
	return &Service{Repository: r, Auth: *NewAuth(r)}
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

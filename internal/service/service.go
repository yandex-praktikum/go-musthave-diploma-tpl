package service

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	numbergenerator "Loyalty/pkg/NumberGenerator"
)

type Repository interface {
	//auth methods
	SaveUser(*models.User, string) error
	GetUser(*models.User) (string, error)
	//user methods
	CreateLoyaltyAccount(string) error
	SaveOrder(order *models.Order, login string) error
	GetOrders(login string) ([]models.Order, error)
	GetBalance(login string) (*models.Account, error)
}

type Service struct {
	Repository
	Auth
}

func NewService(r *repository.Repository) *Service {
	return &Service{Repository: r, Auth: *NewAuth(r)}
}

func (s *Service) CreateLoyaltyAccount(user *models.User) (string, error) {
	//create account number
	number, err := numbergenerator.GenerateNumber(15)
	if err != nil {
		return "", err
	}
	//save accoun in db
	if err := s.Repository.CreateLoyaltyAccount(number); err != nil {
		return "", err
	}
	return number, nil
}

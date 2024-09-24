package service

import (
	"gophermart/internal/models"
	"gophermart/internal/repository"
)

type UserBalanceService struct {
	UserBalanceRepository *repository.UserBalanceRepository
}

func (ubs *UserBalanceService) UpdateUserBalance(accrual float32, userID int) error {
	return ubs.UserBalanceRepository.UpdateUserBalance(accrual, userID)
}

func (ubs *UserBalanceService) GetUserBalance(userID int) (repository.UserBalance, error) {
	return ubs.UserBalanceRepository.GetUserBalance(userID)
}

func (ubs *UserBalanceService) CreateUserBalance(user models.User) error {
	return ubs.UserBalanceRepository.CreateUserBalance(user)
}

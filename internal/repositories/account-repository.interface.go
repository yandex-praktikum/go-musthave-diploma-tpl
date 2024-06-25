package repositories

import "github.com/ShukinDmitriy/gophermart/internal/entities"

type AccountRepositoryInterface interface {
	FindByUserID(userID uint, accountType entities.AccountType) (*entities.Account, error)
}

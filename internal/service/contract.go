package service

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type autorisation interface {
	CreateUser(user models.User) (int, error)
	GetUser(username string) (models.User, error)
}

type balance interface {
	GetBalance(userID int) (models.Balance, error)
	DoWithdraw(userID int, withdraw models.Withdraw) error
	GetWithdraws(userID int) ([]models.WithdrawResponse, error)
}

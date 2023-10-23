package service

import (
	"errors"
	"strconv"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/luhn"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type AccountService struct {
	repo repository.Balance
}

func NewAccountService(repo repository.Balance) *AccountService {
	return &AccountService{repo: repo}
}

func (b *AccountService) GetWithdraws(userID int) ([]models.WithdrawResponse, error) {
	return b.repo.GetWithdraws(userID)
}

func (b *AccountService) GetBalance(userID int) (models.Balance, error) {
	return b.repo.GetBalance(userID)

}
func (b *AccountService) Withdraw(userID int, withdraw models.Withdraw) error {

	numOrderInt, err := strconv.Atoi(withdraw.Order)
	if err != nil {
		return errors.New("PreconditionFailed")
	}

	correctnum := luhn.Valid(numOrderInt)

	if !correctnum {
		return errors.New("UnprocessableEntity")
	}

	err = b.repo.DoWithdraw(userID, withdraw)

	if err != nil {
		return err
	}

	return nil
}

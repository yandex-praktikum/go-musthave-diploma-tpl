package service

import (
	"errors"
	"strconv"

	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/luhn"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/repository"
)

type BalanceService struct {
	repo repository.Balance
}

func NewBalanceStorage(repo repository.Balance) *BalanceService {
	return &BalanceService{repo: repo}
}

func (b *BalanceService) GetWithdraws(userID int) ([]models.WithdrawResponse, error) {
	return b.repo.GetWithdraws(userID)
}

func (b *BalanceService) GetBalance(userID int) (models.Balance, error) {
	return b.repo.GetBalance(userID)

}
func (b *BalanceService) Withdraw(userID int, withdraw models.Withdraw) error {

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

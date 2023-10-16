package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/luhn"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
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

	balance, err := b.repo.GetBalance(userID)

	if err != nil {
		return err
	}

	err = b.repo.DoWithdraw(userID, withdraw)
	fmt.Println("Order ", withdraw.Order, "balance ", balance, withdraw.Sum, "withdraw")

	if err != nil {
		return err
	}

	return nil
}

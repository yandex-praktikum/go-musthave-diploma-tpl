package service

import (
	"github.com/sub3er0/gophermart/internal/repository"
)

type WithdrawService struct {
	WithdrawRepository *repository.WithdrawRepository
}

func (ws *WithdrawService) Withdraw(userID int, orderNumber string, sum int) (int, error) {
	return ws.WithdrawRepository.Withdraw(userID, orderNumber, sum)
}

func (ws *WithdrawService) Withdrawals(userID int) ([]repository.WithdrawInfo, error) {
	return ws.WithdrawRepository.Withdrawals(userID)
}

package balance

import (
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/repository/user/balance"
)

type Service struct {
}

type Balance interface {
	UserBalance()
	BalanceWithdraw()
	UserWithdrawals()
}

func NewService(balance.Repository) Balance {
	return &Service{}
}

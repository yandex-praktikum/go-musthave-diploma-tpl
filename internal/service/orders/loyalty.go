package orders

import (
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

func (s *Service) GetUserByAccessToken(order string, login string, now time.Time) error {
	return s.db.GetUserByAccessToken(order, login, now)
}

func (s *Service) GetAllUserOrders(login string) ([]*models.OrdersUser, error) {
	return s.db.GetAllUserOrders(login)
}

func (s *Service) GetBalanceUser(login string) (*models.Balance, error) {
	return s.db.GetBalanceUser(login)
}

func (s *Service) GetWithdrawals(login string) ([]*models.Withdrawals, error) {
	return s.db.GetWithdrawals(login)
}

func (s *Service) CheckWriteOffOfFunds(login, order string, sum float32, now time.Time) error {
	return s.db.CheckWriteOffOfFunds(login, order, sum, now)
}

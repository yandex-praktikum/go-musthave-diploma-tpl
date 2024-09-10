package service

import (
	"database/sql"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"time"
)

func (s *Service) Get(query string, args ...interface{}) (*sql.Row, error) {
	return s.db.Get(query, args)
}

func (s *Service) GetUserByAccessToken(order string, login string, now time.Time) error {
	return s.db.GetUserByAccessToken(order, login, now)
}

func (s *Service) GetAllUserOrders(login string) ([]*models.OrdersUser, error) {
	return s.db.GetAllUserOrders(login)
}

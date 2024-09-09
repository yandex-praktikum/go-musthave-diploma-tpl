package service

import (
	"database/sql"
	"time"
)

func (s *Service) Get(query string, args ...interface{}) (*sql.Row, error) {
	return s.db.Get(query, args)
}

func (s *Service) GetUserByAccessToken(order string, login string, now time.Time) error {
	return s.db.GetUserByAccessToken(order, login, now)
}

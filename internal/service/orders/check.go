package orders

import (
	"context"
	"time"
)

func (s *Service) CheckTableUserLogin(ctx context.Context, login string) error {
	return s.db.CheckTableUserLogin(ctx, login)
}

func (s *Service) CheckTableUserPassword(ctx context.Context, password string) (string, bool) {
	return s.db.CheckTableUserPassword(ctx, password)
}

func (s *Service) CheckWriteOffOfFunds(ctx context.Context, login, order string, sum float32, now time.Time) error {
	return s.db.CheckWriteOffOfFunds(ctx, login, order, sum, now)
}

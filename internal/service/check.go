package service

import "context"

func (s *Service) CheckTableUserLogin(ctx context.Context, login string) error {
	return s.db.CheckTableUserLogin(ctx, login)
}

func (s *Service) CheckTableUserPassword(ctx context.Context, password string) (string, bool) {
	return s.db.CheckTableUserPassword(ctx, password)
}

package usecase

import (
	"context"

	"github.com/benderr/gophermart/internal/domain/withdrawal"
	"github.com/benderr/gophermart/internal/logger"
)

type withdrawUsecase struct {
	wr     WithdrawRepo
	logger logger.Logger
}

func New(wr WithdrawRepo, l logger.Logger) *withdrawUsecase {
	return &withdrawUsecase{
		wr:     wr,
		logger: l}
}

func (w *withdrawUsecase) GetWithdrawsByUser(ctx context.Context, userid string) ([]withdrawal.Withdrawal, error) {
	return w.wr.GetWithdrawsByUser(ctx, userid)
}

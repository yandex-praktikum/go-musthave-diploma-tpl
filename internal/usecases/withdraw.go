package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"github.com/A-Kuklin/gophermart/internal/storage"
)

type WithdrawUseCases struct {
	strg   storage.WithdrawStorage
	logger logrus.FieldLogger
}

func NewWithdrawUseCases(strg storage.WithdrawStorage, logger logrus.FieldLogger) *WithdrawUseCases {
	return &WithdrawUseCases{
		strg:   strg,
		logger: logger,
	}
}

func (u *WithdrawUseCases) Create(ctx context.Context, withdraw *entities.Withdraw) (*entities.Withdraw, error) {
	withdraw, err := u.strg.CreateWithdraw(ctx, withdraw)
	if err != nil {
		return nil, err
	}

	return withdraw, nil
}

func (u *WithdrawUseCases) GetSumWithdrawals(ctx context.Context, userID uuid.UUID) (int64, error) {
	withdrawals, err := u.strg.GetSumWithdrawals(ctx, userID)
	if err != nil {
		return 0, err
	}

	return withdrawals, nil
}

func (u *WithdrawUseCases) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entities.Withdraw, error) {
	withdrawals, err := u.strg.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

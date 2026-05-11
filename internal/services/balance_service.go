//go:generate mockgen -source=balance_service.go -destination=mocks/mock_balance_service.go -package=mocks
package services

import (
	"context"
	"errors"

	"github.com/KalessinD/gophermart/internal/middleware"
	"github.com/KalessinD/gophermart/internal/models"
	repository "github.com/KalessinD/gophermart/internal/repositories/postgresql"
	"github.com/KalessinD/gophermart/internal/services/alg"
)

type (
	/*
		Объект службы балансов
	*/
	BalanceService struct {
		db repository.SQLStorageInterface
	}

	/*
		Интерфейс оОъекта службы балансов
	*/
	BalanceServiceInterface interface {
		GetBalanceInfo(ctx context.Context) (*models.Balance, error)
		Withdraw(ctx context.Context, withdrawn *models.Withdrawal) error
		ListWithdrawals(ctx context.Context) (models.WithdrawalsList, error)
	}
)

/*
Конструктор службы балансов
*/
func NewBalanceService(db repository.SQLStorageInterface) BalanceServiceInterface {
	return &BalanceService{
		db: db,
	}
}

/*
Получает баланс пользователя
*/
func (s *BalanceService) GetBalanceInfo(ctx context.Context) (*models.Balance, error) {
	claims := middleware.GetClaims(ctx)
	user, err := s.db.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	var withdrawnSum models.Accrual
	withdrawn, err := s.db.GetWithdrawn(ctx, claims.UserID)
	if err != nil {
		if !errors.Is(err, models.ErrWithdrawnNotFound) {
			return nil, err
		}
		withdrawnSum = 0
	} else {
		withdrawnSum = withdrawn.Sum
	}

	return models.NewBalanceItem(user.Balance, withdrawnSum), nil
}

// Списывает с баланса пользователя.
// Может вернуть ошибку
func (s *BalanceService) Withdraw(ctx context.Context, withdrawn *models.Withdrawal) error {
	if !alg.IsValidLuhn(withdrawn.OrderID) {
		return models.ErrOrderWrongFormat
	}

	claims := middleware.GetClaims(ctx)
	withdrawn.UserID = claims.UserID

	return s.db.AddWithdrawn(ctx, withdrawn)
}

func (s *BalanceService) ListWithdrawals(ctx context.Context) (models.WithdrawalsList, error) {
	claims := middleware.GetClaims(ctx)
	list, err := s.db.ListWithdrawals(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	return list, err
}

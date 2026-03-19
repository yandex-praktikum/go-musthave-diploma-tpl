package loyalty

import (
	"context"
	"errors"
	"gmart/internal/adapters/pgc"
	"gmart/internal/config"
	"gmart/internal/domain"
	"gmart/internal/dto"
	"gmart/internal/model/luhn"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=loyalty_mock.go  -package=loyalty

// LoyaltyRepoIface описывает методы репозитория для мокирования
type LoyaltyRepoIface interface {
	GetBalance(ctx context.Context, userID domain.UserID) (current, withdrawn domain.Amount, err error)
	Withdraw(ctx context.Context, userID domain.UserID, order domain.OrderNumber, amount domain.Amount) error
	GetWithdrawals(ctx context.Context, userID domain.UserID) ([]dto.WithdrawalItem, error)
}

// ============ Class ============

type Loyalty struct {
	repo LoyaltyRepoIface
}

// NewLoyalty создает новый объект манипулирования лояльностью
func NewLoyalty(cfg config.Config, pg pgc.PgInstance, m LoyaltyMetrics) (*Loyalty, error) {
	return &Loyalty{
		repo: NewLoyaltyRepo(pg, m),
	}, nil
}

// ============ UseCase ============

// GetBalance возвращает баланс пользователя
func (s *Loyalty) GetBalance(ctx context.Context, userID domain.UserID) (current, withdrawn domain.Amount, err error) {
	return s.repo.GetBalance(ctx, userID)
}

// Withdraw проверяет номер заказа и выполняет списание
func (s *Loyalty) Withdraw(ctx context.Context, userID domain.UserID, order domain.OrderNumber, amount domain.Amount) error {
	// 1. Проверка номера заказа по алгоритму Луна
	if !luhn.IsValid(order) {
		return ErrInvalidOrderNumber
	}

	// 2. Выполнение списания в репозитории
	// Репозиторий сам проверит наличие средств через CONSTRAINT в БД
	return s.repo.Withdraw(ctx, userID, order, amount)
}

// GetWithdrawals возвращает историю списаний
func (s *Loyalty) GetWithdrawals(ctx context.Context, userID domain.UserID) ([]dto.WithdrawalItem, error) {
	return s.repo.GetWithdrawals(ctx, userID)
}

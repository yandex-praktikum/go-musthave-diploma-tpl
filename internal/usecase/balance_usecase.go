package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"github.com/skiphead/go-musthave-diploma-tpl/pkg/utils"
)

// balanceUC реализация BalanceUseCase
type balanceUC struct {
	repo *repository.Repository
}

// NewBalanceUsecase создает новый экземпляр balanceUC
func NewBalanceUsecase(repo *repository.Repository) BalanceUseCase {
	return &balanceUC{
		repo: repo,
	}
}

// GetUserBalance получает баланс пользователя
func (uc *balanceUC) GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	balance, err := uc.repo.User().GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// WithdrawBalance списывает средства с баланса
func (uc *balanceUC) WithdrawBalance(ctx context.Context, userID int, orderNumber string, amount float64) error {
	// Валидация входных параметров
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	if !validateOrderNumber(orderNumber) {
		return ErrInvalidOrderNumber
	}

	// Получаем текущий баланс
	balance, err := uc.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Проверяем достаточность средств
	if balance.Current < amount {
		return ErrInsufficientBalance
	}

	// Начинаем транзакцию
	tx, err := uc.repo.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Гарантируем откат или коммит транзакции
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Проверяем существование заказа
	exists, err := uc.repo.Order().Exists(ctx, orderNumber)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to check order existence: %w", err)
	}

	if exists {
		_ = tx.Rollback(ctx)
		return ErrOrderAlreadyExists
	}

	// Создаем заказ на списание
	order, err := uc.repo.Order().Create(ctx, userID, orderNumber, entity.OrderStatusProcessed)
	if err != nil {
		_ = tx.Rollback(ctx)
		if repository.IsDuplicateError(err) {
			return ErrOrderAlreadyExists
		}
		return fmt.Errorf("failed to create withdrawal order: %w", err)
	}

	// Устанавливаем отрицательную сумму (списание)
	err = uc.repo.Order().UpdateAccrual(ctx, order.ID, -amount, entity.OrderStatusProcessed)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to update withdrawal order: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWithdrawals получает историю списаний
func (uc *balanceUC) GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdrawal, error) {
	withdrawals, err := uc.repo.Withdrawal().GetByUserID(ctx, userID, entity.OrderStatusProcessed)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}

// validateOrderNumber проверяет номер заказа
func validateOrderNumber(number string) bool {
	return utils.IsValidLuhn(number)
}

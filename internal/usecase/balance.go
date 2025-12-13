package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
)

// GetUserBalance получает баланс пользователя
func (uc *useCase) GetUserBalance(ctx context.Context, userID int) (*entity.UserBalance, error) {
	balance, err := uc.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// WithdrawBalance списывает средства с баланса
func (uc *useCase) WithdrawBalance(ctx context.Context, userID int, orderNumber string, amount float64) error {
	// Проверяем валидность номера заказа
	if !validateOrderNumber(orderNumber) {
		return ErrInvalidOrderNumber
	}

	// Проверяем положительность суммы
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
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
	// Откатываем транзакцию только при ошибке
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
		}
	}()

	// Создаем заказ на списание
	order, err := uc.repo.CreateOrder(ctx, userID, orderNumber, entity.OrderStatusProcessed)
	if err != nil {
		if err.Error() == "order already exists for another user" {
			return ErrOrderAlreadyExists
		}
		return fmt.Errorf("failed to create withdrawal order: %w", err)
	}
	// Устанавливаем отрицательную сумму (списание)
	err = uc.repo.UpdateOrderAccrual(ctx, order.ID, -amount, entity.OrderStatusProcessed)
	if err != nil {
		return fmt.Errorf("failed to update withdrawal order: %w", err)
	}

	// Создаем запись о списании (нужно добавить таблицу withdrawals в репозиторий)
	// Это упрощенная версия - в реальности нужно создать отдельную таблицу для списаний

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWithdrawals получает историю списаний
func (uc *useCase) GetWithdrawals(ctx context.Context, userID int) ([]entity.Withdraw, error) {
	// Получаем все заказы пользователя с отрицательным accrual
	orders, err := uc.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	// Фильтруем заказы с отрицательным accrual
	var withdrawals []entity.Withdraw
	for _, order := range orders {
		if order.Accrual != nil && *order.Accrual < 0 {
			withdrawals = append(withdrawals, entity.Withdraw{
				Order:       order.Number,
				Sum:         -*order.Accrual, // Преобразуем отрицательное в положительное
				ProcessedAt: order.UpdatedAt,
			})
		}
	}

	return withdrawals, nil
}

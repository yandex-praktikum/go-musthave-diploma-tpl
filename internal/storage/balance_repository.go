package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ошибки репозитория баланса
var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// BalanceRepository предоставляет методы для работы с балансом и транзакциями
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// NewBalanceRepository создаёт новый репозиторий баланса
func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

// GetBalance получает текущий баланс пользователя
// current - текущая сумма баллов
// withdrawn - сумма использованных баллов за всё время
func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (*UserBalance, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = $2 THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = $3 THEN amount ELSE 0 END), 0) AS current,
			COALESCE(SUM(CASE WHEN type = $3 THEN amount ELSE 0 END), 0) AS withdrawn
		FROM transactions
		WHERE user_id = $1
	`

	var balance UserBalance
	err := r.pool.QueryRow(ctx, query, userID, TransactionTypeCharge, TransactionTypeWithdraw).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)

	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return &balance, nil
}

// AddCharge добавляет транзакцию начисления баллов
func (r *BalanceRepository) AddCharge(ctx context.Context, userID int64, amount float64, orderNumber *string) error {
	query := `
		INSERT INTO transactions (user_id, type, amount, order_number)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, userID, TransactionTypeCharge, amount, orderNumber)
	if err != nil {
		return fmt.Errorf("add charge transaction: %w", err)
	}

	return nil
}

// AddWithdraw добавляет транзакцию списания баллов
// Возвращает ErrInsufficientFunds если недостаточно средств
func (r *BalanceRepository) AddWithdraw(ctx context.Context, userID int64, amount float64, orderNumber *string) error {
	// Проверяем баланс перед списанием
	balance, err := r.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("check balance before withdraw: %w", err)
	}

	if balance.Current < amount {
		return ErrInsufficientFunds
	}

	query := `
		INSERT INTO transactions (user_id, type, amount, order_number)
		VALUES ($1, $2, $3, $4)
	`

	_, err = r.pool.Exec(ctx, query, userID, TransactionTypeWithdraw, amount, orderNumber)
	if err != nil {
		return fmt.Errorf("add withdraw transaction: %w", err)
	}

	return nil
}

// CreateWithdrawal создаёт запись о списании
func (r *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, order_number, sum, processed_at
	`

	var withdrawal Withdrawal
	err := r.pool.QueryRow(ctx, query, userID, orderNumber, sum).Scan(
		&withdrawal.ID,
		&withdrawal.UserID,
		&withdrawal.OrderNumber,
		&withdrawal.Sum,
		&withdrawal.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}

	return nil
}

// GetWithdrawalsByUser получает все списания пользователя, отсортированные по времени (от новых к старым)
func (r *BalanceRepository) GetWithdrawalsByUser(ctx context.Context, userID int64) ([]Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals by user: %w", err)
	}
	defer rows.Close()

	var withdrawals []Withdrawal
	for rows.Next() {
		var withdrawal Withdrawal
		if err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.OrderNumber,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate withdrawals: %w", err)
	}

	return withdrawals, nil
}

// ProcessOrderAccrual обрабатывает начисление по заказу в транзакции
// 1. Обновляет статус заказа
// 2. Добавляет транзакцию начисления
func (r *BalanceRepository) ProcessOrderAccrual(ctx context.Context, orderID int64, userID int64, status string, accrual *float64, orderNumber string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем статус заказа
	updateQuery := `
		UPDATE orders
		SET status = $2, accrual = $3
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, updateQuery, orderID, status, accrual)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// Если есть начисление, добавляем транзакцию
	if accrual != nil && *accrual > 0 {
		chargeQuery := `
			INSERT INTO transactions (user_id, type, amount, order_number)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, chargeQuery, userID, TransactionTypeCharge, *accrual, orderNumber)
		if err != nil {
			return fmt.Errorf("add charge transaction: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
)

// ==================== Реализация WithdrawalRepository ====================

// Create создает новое списание
func (r *withdrawalRepo) Create(ctx context.Context, withdrawal *entity.Withdrawal) error {
	// Проверяем, что сумма отрицательная для списания
	if withdrawal.Sum >= 0 {
		return fmt.Errorf("withdrawal sum must be negative")
	}

	// Создаем заказ со статусом PROCESSED и отрицательной суммой
	query := `
		INSERT INTO orders (number, user_id, status, accrual)
		VALUES ($1, $2, 'PROCESSED', $3)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		withdrawal.Order,
		withdrawal.UserID,
		withdrawal.Sum,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if IsDuplicateError(err) {
			return fmt.Errorf("%w: order with number '%s' already exists", ErrDuplicateKey, withdrawal.Order)
		}
		if IsForeignKeyError(err) {
			return fmt.Errorf("%w: user with ID %d not found", ErrForeignKey, withdrawal.UserID)
		}
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	withdrawal.ProcessedAt = updatedAt.Format(time.RFC3339)
	return nil
}

// GetByUserID получает списания пользователя
func (r *withdrawalRepo) GetByUserID(ctx context.Context, userID int, status entity.OrderStatus) ([]entity.Withdrawal, error) {
	query := `
		SELECT number, accrual, updated_at, user_id
		FROM orders 
		WHERE user_id = $1 
			AND status = $2 
			AND accrual < 0
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query user withdrawals: %w", err)
	}
	defer rows.Close()

	return r.scanWithdrawals(rows)
}

// GetAll получает все списания с пагинацией
func (r *withdrawalRepo) GetAll(ctx context.Context, limit, offset int) ([]entity.Withdrawal, error) {
	query := `
		SELECT number, accrual, updated_at, user_id
		FROM orders 
		WHERE accrual < 0 AND status = 'PROCESSED'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query all withdrawals: %w", err)
	}
	defer rows.Close()

	return r.scanWithdrawals(rows)
}

// GetTotalWithdrawn получает общую сумму списаний пользователя
func (r *withdrawalRepo) GetTotalWithdrawn(ctx context.Context, userID int) (float64, error) {
	query := `
		SELECT COALESCE(ABS(SUM(accrual)), 0)
		FROM orders
		WHERE user_id = $1 
			AND status = 'PROCESSED' 
			AND accrual < 0
	`

	var total float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total withdrawn: %w", err)
	}

	return total, nil
}

// GetWithdrawalByOrder получает списание по номеру заказа
func (r *withdrawalRepo) GetWithdrawalByOrder(ctx context.Context, orderNumber string) (*entity.Withdrawal, error) {
	query := `
		SELECT number, accrual, updated_at, user_id
		FROM orders 
		WHERE number = $1 
			AND accrual < 0 
			AND status = 'PROCESSED'
	`

	withdrawal := &entity.Withdrawal{}
	err := r.scanWithdrawal(
		r.db.QueryRow(ctx, query, orderNumber),
		withdrawal,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: withdrawal with order number '%s'", ErrNotFound, orderNumber)
		}
		return nil, fmt.Errorf("failed to get withdrawal by order: %w", err)
	}

	return withdrawal, nil
}

// GetWithdrawalsSummary получает сводку по списаниям
func (r *withdrawalRepo) GetWithdrawalsSummary(ctx context.Context) (*entity.WithdrawalsSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_withdrawals,
			COALESCE(ABS(SUM(accrual)), 0) as total_amount,
			COUNT(DISTINCT user_id) as unique_users
		FROM orders
		WHERE accrual < 0 AND status = 'PROCESSED'
	`

	summary := &entity.WithdrawalsSummary{}
	err := r.db.QueryRow(ctx, query).Scan(
		&summary.TotalWithdrawals,
		&summary.TotalAmount,
		&summary.UniqueUsers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals summary: %w", err)
	}

	return summary, nil
}

// GetUserWithdrawalsSummary получает сводку по списаниям пользователя
func (r *withdrawalRepo) GetUserWithdrawalsSummary(ctx context.Context, userID int) (*entity.UserWithdrawalsSummary, error) {
	query := `
		SELECT 
			COUNT(*) as withdrawal_count,
			COALESCE(ABS(SUM(accrual)), 0) as total_amount,
			MIN(updated_at) as first_withdrawal,
			MAX(updated_at) as last_withdrawal
		FROM orders
		WHERE user_id = $1 
			AND accrual < 0 
			AND status = 'PROCESSED'
	`

	summary := &entity.UserWithdrawalsSummary{}
	var firstWithdrawal, lastWithdrawal *time.Time

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&summary.WithdrawalCount,
		&summary.TotalAmount,
		&firstWithdrawal,
		&lastWithdrawal,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Если нет списаний, возвращаем пустую сводку
			return &entity.UserWithdrawalsSummary{
				UserID:          userID,
				WithdrawalCount: 0,
				TotalAmount:     0,
				FirstWithdrawal: "",
				LastWithdrawal:  "",
			}, nil
		}
		return nil, fmt.Errorf("failed to get user withdrawals summary: %w", err)
	}

	summary.UserID = userID
	if firstWithdrawal != nil {
		summary.FirstWithdrawal = firstWithdrawal.Format(time.RFC3339)
	}
	if lastWithdrawal != nil {
		summary.LastWithdrawal = lastWithdrawal.Format(time.RFC3339)
	}

	return summary, nil
}

// GetRecentWithdrawals получает последние списания
func (r *withdrawalRepo) GetRecentWithdrawals(ctx context.Context, limit int) ([]entity.Withdrawal, error) {
	query := `
		SELECT number, accrual, updated_at, user_id
		FROM orders 
		WHERE accrual < 0 AND status = 'PROCESSED'
		ORDER BY updated_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent withdrawals: %w", err)
	}
	defer rows.Close()

	return r.scanWithdrawals(rows)
}

// GetWithdrawalsByPeriod получает списания за период
func (r *withdrawalRepo) GetWithdrawalsByPeriod(ctx context.Context, start, end time.Time) ([]entity.Withdrawal, error) {
	query := `
		SELECT number, accrual, updated_at, user_id
		FROM orders 
		WHERE accrual < 0 
			AND status = 'PROCESSED'
			AND updated_at BETWEEN $1 AND $2
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query withdrawals by period: %w", err)
	}
	defer rows.Close()

	return r.scanWithdrawals(rows)
}

// ==================== Вспомогательные методы для WithdrawalRepository ====================

// scanWithdrawal сканирует списание из Row или Rows
func (r *withdrawalRepo) scanWithdrawal(scanner Scanner, withdrawal *entity.Withdrawal) error {
	var updatedAt time.Time

	err := scanner.Scan(
		&withdrawal.Order,
		&withdrawal.Sum,
		&updatedAt,
		&withdrawal.UserID,
	)
	if err != nil {
		return err
	}

	withdrawal.ProcessedAt = updatedAt.Format(time.RFC3339)
	// Преобразуем отрицательную сумму в положительную для отображения
	if withdrawal.Sum < 0 {
		withdrawal.Sum = -withdrawal.Sum
	}

	return nil
}

// scanWithdrawals сканирует список списаний
func (r *withdrawalRepo) scanWithdrawals(rows pgx.Rows) ([]entity.Withdrawal, error) {
	var withdrawals []entity.Withdrawal

	for rows.Next() {
		var withdrawal entity.Withdrawal
		if err := r.scanWithdrawal(rows, &withdrawal); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return withdrawals, nil
}

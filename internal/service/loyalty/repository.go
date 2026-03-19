package loyalty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc"
	"gmart/internal/domain"
	"gmart/internal/dto"

	"github.com/jackc/pgx/v5"
)

// SQL запросы
const (
	// Получаем вычисленный баланс и общую сумму списаний
	sqlGetBalance = `
		SELECT (accrual - withdrawn) AS current, withdrawn 
		FROM balances 
		WHERE user_id = $1
	`

	// Списание баллов. Триггер в БД сам проверит достаточность средств (constraint)
	sqlWithdraw = `
		INSERT INTO withdrawals (user_id, order_number, amount, processed_at)
		VALUES ($1, $2, $3, $4)
	`

	// Список списаний от новых к старым
	sqlGetWithdrawals = `
		SELECT order_number, amount, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
)

var (
	ErrInsufficientFunds = errors.New("insufficient balance")
	ErrWithdrawConflict  = errors.New("withdrawal for this order already exists")
)

type LoyaltyMetrics interface {
	ObserveDB(op metrics.OpType, duration time.Duration)
	IncWithdrawal(status string) // success, insufficient_funds, conflict
	ObserveWithdrawalAmount(amount domain.Amount)
}

type LoyaltyRepo struct {
	pg      pgc.PgInstance
	metrics LoyaltyMetrics
	now     func() time.Time
}

func NewLoyaltyRepo(pg pgc.PgInstance, m LoyaltyMetrics) *LoyaltyRepo {
	return &LoyaltyRepo{
		pg:      pg,
		metrics: m,
		now:     time.Now,
	}
}

// GetBalance возвращает текущий остаток и общую сумму списаний
func (r *LoyaltyRepo) GetBalance(ctx context.Context, userID domain.UserID) (current domain.Amount, withdrawn domain.Amount, err error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpQuery, time.Since(start))
		}
	}()

	err = r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		err := pool.QueryRow(ctx, sqlGetBalance, userID).Scan(&current, &withdrawn)
		if errors.Is(err, pgx.ErrNoRows) {
			current, withdrawn = 0, 0
			return nil
		}
		return err
	})

	if err != nil {
		slog.Error("db query failed", "op", "LoyaltyRepo.GetBalance", "err", err, "user_id", userID)
		return 0, 0, fmt.Errorf("get balance fail: %w", err)
	}

	return current, withdrawn, nil
}

// Withdraw регистрирует списание баллов
func (r *LoyaltyRepo) Withdraw(ctx context.Context, userID domain.UserID, order domain.OrderNumber, amount domain.Amount) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpExec, time.Since(start))
		}
	}()

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		_, err := pool.Exec(ctx, sqlWithdraw, userID, order, amount, r.now())
		return err
	})

	if err != nil {
		// Проверка на отрицательный баланс (сработал CHECK в БД)
		if strings.Contains(err.Error(), "check_balance_not_negative") {
			if r.metrics != nil {
				r.metrics.IncWithdrawal("insufficient_funds")
			}
			return ErrInsufficientFunds
		}
		// Проверка на дубликат номера заказа в списаниях
		if strings.Contains(err.Error(), "withdrawals_order_number_key") {
			if r.metrics != nil {
				r.metrics.IncWithdrawal("conflict")
			}
			return ErrWithdrawConflict
		}

		slog.Error("db exec failed", "op", "LoyaltyRepo.Withdraw", "err", err, "user_id", userID, "order", order)
		return fmt.Errorf("withdraw fail: %w", err)
	}

	if r.metrics != nil {
		r.metrics.IncWithdrawal("success")
		r.metrics.ObserveWithdrawalAmount(amount)
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (r *LoyaltyRepo) GetWithdrawals(ctx context.Context, userID domain.UserID) ([]dto.WithdrawalItem, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpQuery, time.Since(start))
		}
	}()

	result := make([]dto.WithdrawalItem, 0, 10)

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		rows, err := pool.Query(ctx, sqlGetWithdrawals, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item dto.WithdrawalItem
			if err := rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})

	if err != nil {
		slog.Error("db query failed", "op", "LoyaltyRepo.GetWithdrawals", "err", err, "user_id", userID)
		return nil, fmt.Errorf("get withdrawals fail: %w", err)
	}

	return result, nil
}

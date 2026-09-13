package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

// BalanceRepository — репозиторий балансов, работающий с таблицей balances.
type BalanceRepository struct {
	db *sql.DB
}

// NewBalanceRepository создаёт новый экземпляр BalanceRepository поверх переданного соединения с БД.
func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetBalance получает баланс пользователя
func (r *BalanceRepository) GetBalance(userID int64) (*domain.Balance, error) {
	balance := &domain.Balance{UserID: userID}
	query := `
        SELECT current, withdrawn
        FROM balances
        WHERE user_id = $1
    `

	err := r.db.QueryRow(query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Если баланса нет, создаем с нулевыми значениями
			return r.CreateBalance(userID)
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// CreateBalance создает нулевой баланс для пользователя
func (r *BalanceRepository) CreateBalance(userID int64) (*domain.Balance, error) {
	query := `
        INSERT INTO balances (user_id, current, withdrawn)
        VALUES ($1, 0, 0)
        RETURNING current, withdrawn
    `

	balance := &domain.Balance{UserID: userID}
	err := r.db.QueryRow(query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance: %w", err)
	}

	return balance, nil
}

// AddBalance добавляет баллы на счет (начисление)
func (r *BalanceRepository) AddBalance(userID int64, amount float64) error {
	// Используем транзакцию для атомарности
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
        UPDATE balances
        SET current = current + $1
        WHERE user_id = $2
    `

	result, err := tx.Exec(query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Если баланса нет, создаем
		_, err := tx.Exec(`
            INSERT INTO balances (user_id, current, withdrawn)
            VALUES ($1, $2, 0)
        `, userID, amount)
		if err != nil {
			return fmt.Errorf("failed to create balance on add: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SubtractBalance списывает баллы со счета
func (r *BalanceRepository) SubtractBalance(userID int64, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, достаточно ли средств
	var current float64
	checkQuery := `SELECT current FROM balances WHERE user_id = $1 FOR UPDATE`
	err = tx.QueryRow(checkQuery, userID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrInsufficientBalance
		}
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if current < amount {
		return domain.ErrInsufficientBalance
	}

	// Обновляем баланс (уменьшаем current, увеличиваем withdrawn)
	updateQuery := `
        UPDATE balances
        SET current = current - $1, withdrawn = withdrawn + $1
        WHERE user_id = $2
    `
	_, err = tx.Exec(updateQuery, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to subtract balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SubtractBalanceTx списывает баллы со счета в рамках переданной транзакции.
// Возвращает domain.ErrInsufficientBalance, если средств недостаточно.
func (r *BalanceRepository) SubtractBalanceTx(tx *sql.Tx, userID int64, amount float64) error {
	var current float64
	checkQuery := `SELECT current FROM balances WHERE user_id = $1 FOR UPDATE`
	err := tx.QueryRow(checkQuery, userID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrInsufficientBalance
		}
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if current < amount {
		return domain.ErrInsufficientBalance
	}

	updateQuery := `
        UPDATE balances
        SET current = current - $1, withdrawn = withdrawn + $1
        WHERE user_id = $2
    `
	if _, err := tx.Exec(updateQuery, amount, userID); err != nil {
		return fmt.Errorf("failed to subtract balance: %w", err)
	}

	return nil
}

// DB возвращает *sql.DB, чтобы сервис мог управлять транзакциями.
func (r *BalanceRepository) DB() *sql.DB {
	return r.db
}

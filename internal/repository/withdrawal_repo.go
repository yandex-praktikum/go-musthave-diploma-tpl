package repository

import (
	"database/sql"
	"fmt"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

// WithdrawalRepository — репозиторий списаний, работающий с таблицей withdrawals.
type WithdrawalRepository struct {
	db *sql.DB
}

// NewWithdrawalRepository создаёт новый экземпляр WithdrawalRepository поверх переданного соединения с БД.
func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

// Create создает запись о списании
func (r *WithdrawalRepository) Create(withdrawal *domain.Withdrawal) error {
	query := `
        INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	err := r.db.QueryRow(
		query,
		withdrawal.UserID,
		withdrawal.OrderNumber,
		withdrawal.Sum,
		withdrawal.ProcessedAt,
	).Scan(&withdrawal.ID)

	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

// FindByUserID находит все списания пользователя, сортирует по времени (новые сначала)
func (r *WithdrawalRepository) FindByUserID(userID int64) ([]*domain.Withdrawal, error) {
	query := `
        SELECT id, user_id, order_number, sum, processed_at
        FROM withdrawals
        WHERE user_id = $1
        ORDER BY processed_at DESC
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*domain.Withdrawal
	for rows.Next() {
		w := &domain.Withdrawal{}
		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderNumber,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate withdrawals: %w", err)
	}

	return withdrawals, nil
}

// GetTotalWithdrawn возвращает общую сумму списаний пользователя
func (r *WithdrawalRepository) GetTotalWithdrawn(userID int64) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total withdrawn: %w", err)
	}

	return total, nil
}

// CreateTx создает запись о списании в рамках переданной транзакции.
func (r *WithdrawalRepository) CreateTx(tx *sql.Tx, withdrawal *domain.Withdrawal) error {
	query := `
        INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := tx.QueryRow(
		query,
		withdrawal.UserID,
		withdrawal.OrderNumber,
		withdrawal.Sum,
		withdrawal.ProcessedAt,
	).Scan(&withdrawal.ID)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}
	return nil
}

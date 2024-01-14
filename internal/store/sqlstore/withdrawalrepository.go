package sqlstore

import (
	"github.com/iRootPro/gophermart/internal/entity"
	"time"
)

type WithdrawalRepository struct {
	store *Store
}

func (r *WithdrawalRepository) Create(userID int, order string, sum float64) error {
	_, err := r.store.db.Exec("INSERT INTO withdrawal (user_id, order_id, sum, processed_at) VALUES ($1, $2, $3, $4)", userID, order, sum, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (r *WithdrawalRepository) GetByUserID(userID int) ([]*entity.Withdrawal, error) {
	var withdrawals []*entity.Withdrawal
	rows, err := r.store.db.Query("SELECT id, user_id, order_id, sum, processed_at FROM withdrawal WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for rows.Next() {
		withdrawal := &entity.Withdrawal{}
		err := rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderID, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}

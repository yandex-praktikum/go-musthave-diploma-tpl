package sqlstore

import (
	"github.com/iRootPro/gophermart/internal/entity"
	"time"
)

type BalanceRepository struct {
	store *Store
}

func (b *BalanceRepository) Create(userID int) error {
	time := time.Now()
	_, err := b.store.db.Exec("INSERT INTO balance (user_id, current, withdrawn, updated_at) VALUES ($1, $2, $3, $4)", userID, 0, 0, time)
	if err != nil {
		return err
	}
	return nil
}

func (b *BalanceRepository) Get(userID int) (*entity.Balance, error) {
	balance := &entity.Balance{}
	if err := b.store.db.QueryRow("SELECT id, user_id, current, withdrawn, updated_at FROM balance WHERE user_id = $1", userID).Scan(&balance.ID, &balance.UserID, &balance.Current, &balance.Withdrawn, &balance.UpdatedAt); err != nil {
		return nil, err
	}

	return balance, nil
}

func (b *BalanceRepository) UpdateCurrentByUserID(userID int, accrual float64) error {
	_, err := b.store.db.Exec("UPDATE balance SET current = $1 WHERE user_id = $2", accrual, userID)
	if err != nil {
		return err
	}
	return nil
}

func (b *BalanceRepository) UpdateWithdrawnAndCurrentByUserID(userID int, current float64, withdrawn float64) error {
	_, err := b.store.db.Exec("UPDATE balance SET withdrawn = $1, current = $2 WHERE user_id = $3", withdrawn, current, userID)
	if err != nil {
		return err
	}
	return nil
}

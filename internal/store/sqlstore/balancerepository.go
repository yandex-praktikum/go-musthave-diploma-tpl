package sqlstore

import "github.com/iRootPro/gophermart/internal/entity"

type BalanceRepository struct {
	store *Store
}

func (b *BalanceRepository) Get(userID int) (*entity.Balance, error) {
	balance := &entity.Balance{}
	if err := b.store.db.QueryRow("SELECT id, user_id, current, withdrawn, updated_at FROM balance WHERE user_id = $1", userID).Scan(&balance.ID, &balance.UserID, &balance.Current, &balance.Withdrawn, &balance.UpdatedAt); err != nil {
		return nil, err
	}

	return balance, nil
}

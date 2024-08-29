package database

import (
	"context"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

func (db *Database) SelectUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*models.Withdraw, error) {
	query := "SELECT id, \"order\", sum, processed_at, user_id FROM withdrawals WHERE user_id = $1"
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	withdrawals := []*models.Withdraw{}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		var withdraw models.Withdraw
		err = rows.Scan(&withdraw.ID, &withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt, &withdraw.UserID)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdraw)
	}
	return withdrawals, nil
}

func (db *Database) InsertWithdraw(ctx context.Context, withdraw *models.Withdraw) error {
	query := "INSERT INTO withdrawals (id, \"order\", sum, processed_at, user_id) VALUES($1,$2,$3,$4,$5)"
	_, err := db.Exec(ctx, query, withdraw.ID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt, withdraw.UserID)
	if err != nil {
		return err
	}
	return nil
}

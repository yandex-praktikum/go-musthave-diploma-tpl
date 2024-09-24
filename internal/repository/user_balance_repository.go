package repository

import (
	"gophermart/internal/models"
	"gophermart/storage"
)

type UserBalanceRepository struct {
	DBStorage *storage.PgStorage
}

func (ubr *UserBalanceRepository) UpdateUserBalance(accrual float32, userID int) error {
	query := "UPDATE user_balance SET current = current + $1 WHERE user_id = $2"
	_, err := ubr.DBStorage.Conn.Exec(ubr.DBStorage.Ctx, query, accrual, userID)

	return err
}

func (ubr *UserBalanceRepository) GetUserBalance(userID int) (UserBalance, error) {
	var userBalance UserBalance

	query := "SELECT current, withdrawn FROM user_balance WHERE user_id = $1"
	rows, err := ubr.DBStorage.Conn.Query(ubr.DBStorage.Ctx, query, userID)

	if err != nil {
		return userBalance, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&userBalance.Current, &userBalance.Withdrawn); err != nil {
			return userBalance, err
		}
	}

	if err := rows.Err(); err != nil {
		return userBalance, err
	}

	return userBalance, nil
}

func (ubr *UserBalanceRepository) CreateUserBalance(user models.User) error {
	query := "INSERT INTO user_balance (user_id, current) VALUES ($1, 0)"
	_, err := ubr.DBStorage.Conn.Exec(ubr.DBStorage.Ctx, query, user.ID)

	return err
}

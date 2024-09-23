package repository

import "gophermart/db"

type UserBalanceRepository struct {
	DBStorage *db.PgStorage
}

func (ubr *UserBalanceRepository) UpdateUserBalance(accrual float32, userID int) error {
	query := "UPDATE user_balance SET balance = balance + $1 WHERE user_id = $2"
	_, err := ubr.DBStorage.Conn.Exec(ubr.DBStorage.Ctx, query, accrual, userID)

	return err
}

func (ubr *UserBalanceRepository) GetUserBalance(userID int) (UserBalance, error) {
	var userBalance UserBalance

	query := "SELECT balance, used_balance FROM user_balance WHERE user_id = $1"
	rows, err := ubr.DBStorage.Conn.Query(ubr.DBStorage.Ctx, query, userID)

	if err != nil {
		return userBalance, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&userBalance.Balance, &userBalance.UsedBalance); err != nil {
			return userBalance, err
		}
	}

	if err := rows.Err(); err != nil {
		return userBalance, err
	}

	return userBalance, nil
}

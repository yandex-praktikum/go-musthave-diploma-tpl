package repository

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"gophermart/db"
	"time"
)

type WithdrawRepository struct {
	DBStorage *db.PgStorage
}

type WithdrawInfo struct {
	OrderNumber string    `json:"number"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (wr *WithdrawRepository) Withdraw(userID int, orderNumber string, sum int) (int, error) {
	var userBalance int
	tx, err := wr.DBStorage.Conn.BeginTx(wr.DBStorage.Ctx, pgx.TxOptions{})

	if err != nil {
		return -1, fmt.Errorf("ошибка при начале транзакции: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(wr.DBStorage.Ctx)
		}
	}()

	query := "SELECT balance FROM user_balance WHERE user_id = $1"
	err = tx.QueryRow(wr.DBStorage.Ctx, query, userID).Scan(&userBalance)

	if err != nil {
		return -1, err
	}

	if userBalance < sum {
		return -2, nil
	}

	query = "UPDATE user_balance SET balance = balance - $1, used_balance = used_balance + $1 WHERE user_id = $2"
	_, err = tx.Exec(wr.DBStorage.Ctx, query, sum, userID)
	if err != nil {
		tx.Rollback(wr.DBStorage.Ctx)
		return -1, err
	}

	query = "INSERT INTO withdrawal (user_id, order_number, sum) VALUES ($1, $2, $3)"
	_, err = tx.Exec(wr.DBStorage.Ctx, query, userID, orderNumber, sum)
	if err != nil {
		tx.Rollback(wr.DBStorage.Ctx)
		return -1, err
	}

	if err := tx.Commit(wr.DBStorage.Ctx); err != nil {
		return -1, err
	}

	return 0, nil
}

func (wr *WithdrawRepository) Withdrawals(userID int) ([]WithdrawInfo, error) {
	var withdrawalInfoArray []WithdrawInfo

	query := "SELECT order_number, sum, created_at FROM withdrawal WHERE user_id = $1 ORDER BY created_at DESC"
	rows, err := wr.DBStorage.Conn.Query(wr.DBStorage.Ctx, query, userID)

	if err != nil {
		return withdrawalInfoArray, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawalInfo WithdrawInfo
		if err := rows.Scan(&withdrawalInfo.OrderNumber, &withdrawalInfo.Sum, &withdrawalInfo.ProcessedAt); err != nil {
			return withdrawalInfoArray, err
		}

		withdrawalInfoArray = append(withdrawalInfoArray, withdrawalInfo)
	}

	if err := rows.Err(); err != nil {
		return withdrawalInfoArray, err
	}

	return withdrawalInfoArray, nil
}

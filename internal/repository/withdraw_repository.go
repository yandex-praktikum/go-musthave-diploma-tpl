package repository

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"gophermart/storage"
	"time"
)

type WithdrawRepository struct {
	DBStorage *storage.PgStorage
}

type WithdrawInfo struct {
	OrderNumber string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

func (wr *WithdrawRepository) Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error) {
	var userBalance decimal.Decimal
	tx, err := wr.DBStorage.Conn.BeginTx(wr.DBStorage.Ctx, pgx.TxOptions{})

	if err != nil {
		return WithdrawTransactionError, fmt.Errorf("ошибка при начале транзакции: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(wr.DBStorage.Ctx)
		}
	}()

	query := "SELECT current FROM user_balance WHERE user_id = $1"
	err = tx.QueryRow(wr.DBStorage.Ctx, query, userID).Scan(&userBalance)

	if err != nil {
		return WithdrawTransactionError, err
	}

	if userBalance < sum {
		return -2, nil
	}

	query = "UPDATE user_balance SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2"
	_, err = tx.Exec(wr.DBStorage.Ctx, query, sum, userID)
	if err != nil {
		tx.Rollback(wr.DBStorage.Ctx)
		return -1, err
	}

	currentTime := time.Now()
	query = "INSERT INTO withdrawal (user_id, order_number, sum, created_at) VALUES ($1, $2, $3, $4)"
	_, err = tx.Exec(wr.DBStorage.Ctx, query, userID, orderNumber, sum, currentTime)
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

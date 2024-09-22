package repository

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type BalancePostgres struct {
	db *sqlx.DB
}

func NewBalancePostgres(db *sqlx.DB) *BalancePostgres {
	return &BalancePostgres{db: db}
}

func (b *BalancePostgres) GetBalance(userID int) (models.Balance, error) {

	var balance models.Balance

	query := `SELECT SUM (sum) current,
                SUM(CASE WHEN  sum < 0 THEN -sum ELSE 0 END) withdrawn
                        FROM balance WHERE user_id=$1 group by user_id`

	err := b.db.Get(&balance, query, userID)

	if err != nil {
		return balance, err
	}

	return balance, nil
}

func (b *BalancePostgres) GetWithdraws(userID int) ([]models.WithdrawResponse, error) {
	var withdraws []models.WithdrawResponse

	query := `SELECT number, -sum AS sum, processed from balance WHERE user_id = $1 AND sum < 0`

	err := b.db.Select(&withdraws, query, userID)

	if err != nil {
		return withdraws, err
	}
	return withdraws, nil
}

func (b *BalancePostgres) DoWithdraw(userID int, withdraw models.Withdraw) error {
	var balance float64
	var login string

	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			return
		}
	}()

	stmtLock, err := tx.Prepare(`SELECT login FROM users WHERE id = $1 FOR UPDATE`)

	if err != nil {
		return err
	}
	defer stmtLock.Close()

	stmtBalance, err := tx.Prepare(`SELECT SUM(sum) - $1 AS balance from balance WHERE user_id = $2 group by user_id`)

	if err != nil {
		return err
	}

	defer stmtBalance.Close()

	smtWithdraw, err := tx.Prepare(`INSERT INTO balance (number, user_id, sum) values ($1, $2, $3)`)

	if err != nil {
		return err
	}
	defer smtWithdraw.Close()

	stmtUnLock, err := tx.Prepare(`UPDATE users SET login = $1 WHERE id = $2`)

	if err != nil {
		return err
	}

	defer stmtUnLock.Close()

	err = stmtLock.QueryRow(userID).Scan(&login)
	if err != nil {
		return err
	}
	err = stmtBalance.QueryRow(withdraw.Sum, userID).Scan(&balance)
	if err != nil {
		return err
	}

	_, err = smtWithdraw.Exec(withdraw.Order, userID, -withdraw.Sum)
	if err != nil {
		return err
	}

	_, err = stmtUnLock.Exec(login, userID)
	if err != nil {
		return err
	}

	if balance > 0 {
		err = tx.Commit()
		if err != nil {
			return err
		}
	} else {
		return errors.New("PaymentRequired")
	}

	return nil
}

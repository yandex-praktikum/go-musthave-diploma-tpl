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

	query := `SELECT user_id, SUM (sum) accrual FROM balance WHERE user_id=$1 group by user_id`

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
	var balance int64
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queryB := `SELECT SUM(sum) - $1 AS balance from balance WHERE user_id = $2 group by user_id`
	err = tx.QueryRow(queryB, withdraw.Sum, userID).Scan(&balance)
	if err != nil {
		return err
	}

	if balance < 0 {
		return errors.New("PaymentRequired")
	}

	query := `INSERT INTO balance (number, user_id, sum) values ($1, $2, $3)`
	_, err = b.db.Exec(query, withdraw.Order, userID, -withdraw.Sum)

	if err != nil {
		return err
	}

	_ = tx.Commit()

	return nil
}

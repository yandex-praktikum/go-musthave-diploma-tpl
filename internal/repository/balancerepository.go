package repository

import (
	"database/sql"
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
	var idRow int
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

	query := `INSERT INTO balance (number, user_id, sum)
                SELECT $1, $2, $3 WHERE not exists (SELECT SUM(sum) - $4 AS balance from balance
                        WHERE user_id = $5  group by user_id HAVING SUM(sum) - $6 <= 0) returning id `

	err = tx.QueryRow(query, withdraw.Order, userID, -withdraw.Sum, withdraw.Sum, userID, withdraw.Sum).Scan(&idRow)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("PaymentRequired")
		} else {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

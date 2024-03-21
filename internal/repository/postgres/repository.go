package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

const (
	success                 = `SUCCESS`
	insertWithdrawls        = `INSERT INTO withdrawals (id, user_id, order_number, amount) VALUES ($1, $2, $3, $4)`
	getBalanceStmt          = `SELECT balance FROM balances WHERE user_id = $1`
	updateBalanceStmt       = `UPDATE balances SET balance = $1 WHERE user_id = $2`
	updateWithdrawStmt      = `UPDATE withdraw_balances SET amount = (SELECT amount + $1 FROM withdraw_balances WHERE user_id = $2) WHERE user_id = $2`
	setStatusWithdrawlsStmt = `UPDATE withdrawls SET status = $1 WHERE order_number = $2`
)

var ErrNotEnoughFunds = errors.New("not enough balance")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithdrawalRequest(userID uuid.UUID, orderNumber string, amount float64) error {
	withdrawID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(insertWithdrawls, withdrawID, userID, orderNumber, amount); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance float64

	err = tx.QueryRow(getBalanceStmt, userID).Scan(&balance)
	if err != nil {
		return err
	}

	balance -= amount

	if balance < 0 {
		return ErrNotEnoughFunds
	}

	_, err = tx.Exec(updateBalanceStmt, balance, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(updateWithdrawStmt, amount, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(setStatusWithdrawlsStmt, success, orderNumber)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

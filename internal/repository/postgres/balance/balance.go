package balance

import (
	"context"
	"database/sql"
	"errors"
)

type BalanceRepository interface {
	CreateBalance(ctx context.Context, login string) error
	AddBonus(ctx context.Context, login string, amount float64) error
	WithdrawBonus(ctx context.Context, login string, amount float64) error
	GetBalance(ctx context.Context, login string) (*Balance, error)
}

type Balance struct {
	Login          string
	BonusAmount    float64
	WithDrawAmount float64
}

var ErrNotEnoughFunds = errors.New("not enough funds on balance")

type BalanceStorageDB struct {
	db *sql.DB
}

func New(db *sql.DB) BalanceRepository {
	initDB(db)
	return &BalanceStorageDB{db: db}
}

func (b *BalanceStorageDB) CreateBalance(ctx context.Context, login string) error {

	_, err := b.db.ExecContext(ctx,
		`INSERT INTO balance (login, balance_amount, withdraw_amount) VALUES($1, $2, $3)`, login, 0, 0)
	return err
}

func (b *BalanceStorageDB) AddBonus(ctx context.Context, login string, amount float64) error {
	var (
		updateStmt = `UPDATE balance SET balance_amount = balance_amount + $1 WHERE login = $2`
	)

	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, updateStmt, amount, login)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return err
}

func (b *BalanceStorageDB) WithdrawBonus(ctx context.Context, login string, amount float64) error {
	var (
		selectStmt   = `SELECT balance_amount FROM balance WHERE login = $1`
		updateStmt   = `UPDATE balance SET balance_amount = balance_amount - $1, withdraw_amount = withdraw_amount + $2 WHERE login = $3`
		actualAmount float64
	)

	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, selectStmt, login)

	err = row.Scan(&actualAmount)
	if err != nil {
		return err
	}

	if actualAmount < amount {
		return ErrNotEnoughFunds
	}

	_, err = tx.ExecContext(ctx, updateStmt, amount, amount, login)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return err
}

func (b *BalanceStorageDB) GetBalance(ctx context.Context, login string) (*Balance, error) {
	var (
		selectStmt = `SELECT login, balance_amount, withdraw_amount
		FROM balance WHERE login = $1`
		balance Balance
	)

	row := b.db.QueryRowContext(ctx, selectStmt, login)

	err := row.Scan(&balance.Login, &balance.BonusAmount, &balance.WithDrawAmount)

	return &balance, err
}

func initDB(db *sql.DB) {
	db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXIST balance (
			login varchar
			balance_amount numeric(15, 2),
			withdraw_amount numeric(15,2)
		)`)
	db.ExecContext(context.Background(),
		`CREATE UNIQUE INDEX IF NOT EXIST login_balance_idx ON balance (login)`)
}

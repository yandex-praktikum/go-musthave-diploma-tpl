package withdraw

import (
	"context"
	"database/sql"

	"github.com/SmoothWay/gophermart/internal/model"
)

type WithdrawRepository interface {
	Save(tx *sql.Tx, withdraw *model.Withdraw) error
	FindByUser(login string) ([]model.Withdraw, error)
	FindSumByUser(login string) (float64, error)
}

type withdrawStorageDB struct {
	db *sql.DB
}

func New(db *sql.DB) WithdrawRepository {
	initDB(db)
	return &withdrawStorageDB{db: db}
}

func (w *withdrawStorageDB) Save(tx *sql.Tx, withdraw *model.Withdraw) error {
	_, err := w.db.ExecContext(context.Background(),
		`INSERT INTO withdraws(withdraw_order, login, withdraw_sum, processed_at)
	VALUES($1, $2, $3, $4)`, withdraw.Order, withdraw.Login, withdraw.Sum, withdraw.ProcessedAt)
	return err
}

func (w *withdrawStorageDB) FindByUser(login string) ([]model.Withdraw, error) {
	rows, err := w.db.QueryContext(context.Background(),
		`SELECT withdraw_order, login, withdraw_sum, processed_at
	FROM withdraws
	WHERE login = $1`, login)
	if err != nil {
		return nil, err
	}

	var withdraws []model.Withdraw

	for rows.Next() {
		var row model.Withdraw
		err = rows.Scan(&row.Order, &row.Login, &row.Sum, &row.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdraws = append(withdraws, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdraws, err
}

func (w *withdrawStorageDB) FindSumByUser(login string) (float64, error) {
	row := w.db.QueryRowContext(context.Background(),
		`SELECT withdraw_sum
	FROM withdraws
	WHERE login = $1`, login)

	var sum float64

	err := row.Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, err

}

func initDB(db *sql.DB) {
	db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXIST withdraws (
			withdraw_order VARCHAR.
			login VARCHAR,
			withdraw_sum NUMERIC(15, 2),
			processed_at timestamp
		)`)
	db.ExecContext(context.Background(),
		`CREATE INDEX IF NOT EXIST login_withdraws_idx ON withdraws(login)`)
}

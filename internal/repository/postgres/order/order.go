package order

import (
	"context"
	"database/sql"

	"github.com/SmoothWay/gophermart/internal/model"
)

type OrderRepository interface {
	Save(order *model.Order) error
	FindByUser(login string) ([]model.Order, error)
	FindByNumber(number string) (*model.Order, error)
	FindNumbersToProcess() ([]model.Order, error)
	Update(tx *sql.Tx, order *model.Order) error
}

type orderStorageDB struct {
	db *sql.DB
}

func New(db *sql.DB) OrderRepository {
	initDB(db)
	return &orderStorageDB{
		db: db,
	}
}

func (o *orderStorageDB) Save(order *model.Order) error {
	_, err := o.db.ExecContext(context.Background(),
		`INSERT INTO orders (order_number, login, status, accrual, updated_at)
	VALUES ($1, $2, $3, $4, $5)`, order.Number, order.Login, order.Status, order.Accrual, order.UploadAt)

	return err
}

func (o *orderStorageDB) FindByUser(login string) ([]model.Order, error) {
	rows, err := o.db.QueryContext(context.Background(),
		`SELECT order_number, login, status, accrual, updated_at
	FROM orders
	WHERE login = $1`, login)
	if err != nil {
		return nil, err
	}

	orders := make([]model.Order, 0)

	for rows.Next() {
		var row model.Order

		err = rows.Scan(&row.Number, &row.Login, &row.Status, &row.Accrual, &row.UploadAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *orderStorageDB) FindByNumber(number string) (*model.Order, error) {
	rows, err := o.db.QueryContext(context.Background(),
		`SELECT order_number, login, status, accrual, updated_at
	FROM orders
	WHERE order_number = $1`, number)
	if err != nil {
		return nil, err
	}

	var noRows = true
	var row model.Order
	for rows.Next() {
		noRows = false
		err = rows.Scan(&row.Number, &row.Login, &row.Status, &row.Accrual, &row.UploadAt)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	if noRows {
		return nil, nil
	}
	return &row, nil
}

func (o *orderStorageDB) FindNumbersToProcess() ([]model.Order, error) {
	rows, err := o.db.QueryContext(context.Background(), `
        SELECT order_number, login, status, accrual, o.uploaded_at 
		FROM orders
		WHERE status in (1,2)`)
	if err != nil {
		return nil, err
	}

	orders := make([]model.Order, 0)
	for rows.Next() {
		var row model.Order
		err = rows.Scan(&row.Number, &row.Login, &row.Status, &row.Accrual, &row.UploadAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, row)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (o *orderStorageDB) Update(tx *sql.Tx, order *model.Order) error {
	var (
		err        error
		updateStmt = `UPDATE orders SET status = $1, accrual = $2
						WHERE order_number = $3`
	)
	if tx != nil {
		_, err = tx.ExecContext(context.Background(),
			updateStmt, order.Status, order.Accrual, order.Number)
	} else {
		_, err = o.db.ExecContext(context.Background(),
			updateStmt, order.Status, order.Accrual, order.Number)
	}
	return err
}

func initDB(db *sql.DB) {
	db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXIST orders (
			order_number VARCHAR PRIMARY KEY,
			login VARCHAR
			status INT
			accrual NUMERIC(15, 2)
			uploaded_at TIMESTAMP)
			`)

	db.ExecContext(context.Background(),
		`CREATE UNIQUE INDEX IF NOT EXIST order_number_tx ON orders(order_number)`)

}

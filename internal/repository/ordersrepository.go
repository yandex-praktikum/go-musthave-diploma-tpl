package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

const (
	statusNew       = "NEW"
	statusProcessed = "PROCESSED"
)

type OrdersPostgres struct {
	db *sqlx.DB
}

func NewOrdersPostgres(db *sqlx.DB) *OrdersPostgres {
	return &OrdersPostgres{db: db}
}

func (o *OrdersPostgres) CreateOrder(currentuserID int, num, status string) (int, time.Time, error) {
	var userID int
	var updatedate time.Time

	tx, err := o.db.Begin()
	if err != nil {
		return 0, updatedate, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmtOrd, err := tx.Prepare(
		`INSERT INTO orders (user_id, number, status, updatedate) values ($1, $2, $3, $4)
            ON CONFLICT (number) DO UPDATE SET number =  EXCLUDED.number, updatedate = now() returning user_id, updatedate`)

	if err != nil {
		return 0, updatedate, err
	}
	defer stmtOrd.Close()

	stmtBal, err := tx.Prepare(`INSERT INTO balance (number, user_id, sum) values ($1, $2, 0)`)

	if err != nil {
		return 0, updatedate, err
	}
	defer stmtBal.Close()

	row := stmtOrd.QueryRow(currentuserID, num, status, updatedate)
	if err := row.Scan(&userID, &updatedate); err != nil {
		return 0, updatedate, err
	}
	_, err = stmtBal.Exec(num, currentuserID)
	if err != nil {
		return 0, updatedate, err
	}

	_ = tx.Commit()
	return userID, updatedate, nil
}

func (o *OrdersPostgres) ChangeStatusAndSum(sum float64, status, num string) error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queryO := `UPDATE orders SET status = $1 WHERE number = $2`
	_, err = tx.Exec(queryO, status, num)
	if err != nil {
		return err
	}
	queryB := `UPDATE  balance SET sum = $1 WHERE sum = 0 AND number = $2`
	_, err = tx.Exec(queryB, sum, num)
	if err != nil {
		return err
	}

	_ = tx.Commit()
	return err
}

func (o *OrdersPostgres) GetOrdersWithStatus() ([]models.OrderResponse, error) {
	var lists []models.OrderResponse

	query := `SELECT number, status from orders WHERE status in ($1, $2)`

	err := o.db.Select(&lists, query, statusNew, statusProcessed)

	return lists, err
}

func (o *OrdersPostgres) GetOrders(userID int) ([]models.Order, error) {
	orders := make([]models.Order, 0)
	query := "SELECT  o.number, o.status, b.sum, o.uploaddate FROM ORDERS o LEFT JOIN BALANCE b  ON o.number = b.number WHERE o.user_id = $1 AND b.sum >= 0 ORDER by uploaddate"

	err := o.db.Select(&orders, query, userID)

	if err != nil {
		return orders, err
	}

	return orders, nil
}

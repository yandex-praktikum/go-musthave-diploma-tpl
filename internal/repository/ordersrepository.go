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

func (o *OrdersPostgres) CreateOrder(user_id int, num, status string) (int, time.Time, error) {
	var userId int
	var apdatedate time.Time
	query := `INSERT INTO orders (number, status, user_id, apdatedate) values ($1, $2, $3, $4)
                                ON CONFLICT (number) DO UPDATE SET number =  EXCLUDED.number, apdatedate = now() returning user_id, apdatedate`
	row := o.db.QueryRow(query, num, status, user_id, apdatedate)

	if err := row.Scan(&userId, &apdatedate); err != nil {
		return 0, apdatedate, err
	}

	return userId, apdatedate, nil
}

func (o *OrdersPostgres) ChangeStatusAndSum(sum float64, status, num string) error {

	query := `UPDATE orders SET sum = $1, status = $2 WHERE number = $3`

	_, err := o.db.Exec(query, sum, num, status)
	return err
}

func (o *OrdersPostgres) GetOrdersWithStatus() ([]models.OrderResponse, error) {
	var lists []models.OrderResponse

	query := `SELECT number, status, sum as accrual from orders WHERE status in ($1, $2)`

	err := o.db.Select(&lists, query, statusNew, statusProcessed)

	return lists, err
}

func (o *OrdersPostgres) GetOrders(user_id int) ([]models.Order, error) {
	orders := make([]models.Order, 0)
	query := "SELECT  number, status, sum, uploaddate FROM ORDERS WHERE user_id = $1 Order by uploaddate"

	err := o.db.Select(&orders, query, user_id)

	if err != nil {
		return orders, err
	}

	return orders, nil
}

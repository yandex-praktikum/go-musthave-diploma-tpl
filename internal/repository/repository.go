package repository

import (
	"Loyalty/internal/models"
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type Repository struct {
	db DB
}

func NewRepository(db DB) *Repository {
	return &Repository{db: db}
}

//get user from db
func (r *Repository) CreateLoyaltyAccount(number string) error {
	var i int
	q := `INSERT INTO accounts (number,current,withdrawn)
	VALUES ($1,0,0)
	RETURNING id;`
	r.db.QueryRow(context.Background(), q, number).Scan(&i)
	if i == 0 {
		return ErrInt
	}
	logrus.Printf("created new account %s", number)

	return nil
}

//save order
func (r *Repository) SaveOrder(order *models.Order, login string) error {
	var i int
	var loginFromDb string
	var timeFromDb time.Time
	timeCreated := time.Now()

	q := `INSERT INTO orders(number,user_id,status,accrual,uploaded_at)
	VALUES ($1,(SELECT id FROM users WHERE login=$2),$3,$4,$5)
	ON CONFLICT (number) DO UPDATE SET
	number=EXCLUDED.number
	RETURNING id,uploaded_at,(SELECT login FROM users WHERE id=orders.user_id);`

	r.db.QueryRow(context.Background(), q, order.Number, login, order.Status, order.Accrual, timeCreated).Scan(&i, &timeFromDb, &loginFromDb)
	if i == 0 {
		return ErrInt
	}
	if timeCreated.Unix() != timeFromDb.Unix() {
		if loginFromDb != login {
			return ErrOrdUsrConfl
		}
		return ErrOrdOverLap
	}
	return nil
}

//get orders list
func (r *Repository) GetOrders(login string) ([]models.Order, error) {
	q := `SELECT number, status, accrual, uploaded_at
	FROM orders
		WHERE
	user_id=(SELECT id FROM users WHERE login=$1);`

	rows, err := r.db.Query(context.Background(), q, login)
	if err != nil {
		logrus.Error(err)
		return nil, ErrInt
	}
	var list = make([]models.Order, 0, 10)
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			logrus.Error(err)
			return nil, ErrInt
		}
		list = append(list, order)
	}
	return list, nil
}

//get customer balance
func (r *Repository) GetBalance(login string) (*models.Account, error) {
	q := `SELECT current, withdrawn
	FROM accounts 
		WHERE
	id=(SELECT id FROM users WHERE login=$1);`

	rows, err := r.db.Query(context.Background(), q, login)
	if err != nil {
		logrus.Error(err)
		return nil, ErrInt
	}
	var account models.Account
	for rows.Next() {
		err := rows.Scan(&account.Current, &account.Withdrawn)
		if err != nil {
			logrus.Error(err)
			return nil, ErrInt
		}

	}
	return &account, nil
}

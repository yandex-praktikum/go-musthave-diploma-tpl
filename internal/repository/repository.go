package repository

import (
	"Loyalty/internal/models"
	"context"
	"errors"
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
	q := `INSERT INTO accounts (number,current,withdraw)
	VALUES ($1,0,0)
	RETURNING id;`
	r.db.QueryRow(context.Background(), q, number).Scan(&i)
	if i == 0 {
		return errors.New("error: internal db error")
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
		return errors.New("error: internal db error")
	}
	if timeCreated.Unix() != timeFromDb.Unix() {
		if loginFromDb != login {
			return errors.New("error: order was added by other customer")
		}
		return errors.New("error: order already exist")
	}
	return nil
}

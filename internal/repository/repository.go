package repository

import (
	"Loyalty/internal/models"
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	db    DB
	queue []string
	mx    *sync.Mutex
	cash  *sync.Map
}

func NewRepository(db DB) *Repository {
	return &Repository{
		db:    db,
		queue: make([]string, 0, 10),
		mx:    &sync.Mutex{},
		cash:  &sync.Map{},
	}
}

//get user from db ========================================================
func (r *Repository) CreateLoyaltyAccount(number uint64) error {
	var i int

	q := `INSERT INTO accounts (number, current, withdrawn)
	VALUES ($1,0,0)
	RETURNING id;`
	r.db.QueryRow(context.Background(), q, number).Scan(&i)
	if i == 0 {
		return ErrInt
	}
	logrus.Printf("created new account %d", number)

	return nil
}

//save order ========================================================
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

	row := r.db.QueryRow(context.Background(), q, order.Number, login, order.Status, order.Accrual, timeCreated)
	if err := row.Scan(&i, &timeFromDb, &loginFromDb); err != nil {
		logrus.Error(err)
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

//update order
func (r *Repository) UpdateOrder(order *models.Order) error {
	return nil
}

//get orders list ========================================================
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

//get customer balance ========================================================
func (r *Repository) GetBalance(login string) (*models.Account, error) {
	q := `SELECT current, withdrawn
	FROM accounts 
		WHERE
	id=(SELECT account_id FROM users WHERE login=$1);`

	var account models.Account
	row := r.db.QueryRow(context.Background(), q, login)
	if err := row.Scan(&account.Current, &account.Withdrawn); err != nil {
		logrus.Error(err)
		return nil, ErrInt
	}
	return &account, nil
}

//check order ========================================================
func (r *Repository) CheckOrder(number string, login string) (string, error) {
	var status string
	q := `SELECT status
	FROM orders
		WHERE
	user_id=(SELECT id FROM users WHERE login=$1) and number=$2;`

	res := r.db.QueryRow(context.Background(), q, login, number)
	if err := res.Scan(&status); err != nil {
		logrus.Error(err)
		return "", ErrInt
	}
	return status, nil
}

//withdraw ========================================================
func (r *Repository) Withdraw(withdraw *models.Withdraw, login string) error {
	tx, err := r.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		logrus.Error(err)
		return ErrInt
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()
	var id int
	q := `INSERT INTO withdrawls
	 (order_id, sum, processed_at)
	 	VALUES ((
			 SELECT id FROM orders
		WHERE
			number=$1
		 ),$2,$3)
		 RETURNING id;`
	res := r.db.QueryRow(context.Background(), q, withdraw.Order, withdraw.Sum, time.Now())
	if err := res.Scan(&id); err != nil {
		logrus.Error(err)
		return ErrInt
	}

	q = `UPDATE accounts 
	SET current=current-$1,withdrawn=withdrawn+$1
		WHERE id=
	(SELECT account_id FROM users
		WHERE login=$2);`
	_, err = r.db.Exec(context.Background(), q, withdraw.Sum, login)
	if err != nil {
		logrus.Error(err)
		return ErrInt
	}

	return nil
}

//get withdrawls ========================================================
func (r *Repository) GetWithdrawls(login string) ([]models.Withdraw, error) {
	q := `SELECT number,sum,processed_at
	FROM withdrawls    
	JOIN 
		orders ON order_id=orders.id
	WHERE
		user_id=(SELECT id FROM users
	WHERE login=$1)
		ORDER BY processed_at;`

	rows, err := r.db.Query(context.Background(), q, login)
	if err != nil {
		logrus.Error(err)
		return nil, ErrInt
	}
	var list = make([]models.Withdraw, 0, 10)
	for rows.Next() {
		var withdraw models.Withdraw
		err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.Processed_at)
		if err != nil {
			logrus.Error(err)
			return nil, ErrInt
		}
		list = append(list, withdraw)
	}
	return list, nil
}

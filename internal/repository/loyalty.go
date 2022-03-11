package repository

import (
	"Loyalty/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type loyalty struct {
	db     DB
	logger *logrus.Logger
}

func NewLoyalty(db DB, logger *logrus.Logger) *loyalty {
	return &loyalty{
		db:     db,
		logger: logger,
	}
}

//create account ========================================================
func (l *loyalty) CreateLoyaltyAccount(number uint64) error {
	var i int

	q := `INSERT INTO accounts (number, current, withdrawn)
	VALUES ($1,0,0)
	RETURNING id;`
	l.db.QueryRow(context.Background(), q, number).Scan(&i)
	if i == 0 {
		return ErrInt
	}
	l.logger.Printf("created new account %d", number)

	return nil
}

//save order ========================================================
func (l *loyalty) SaveOrder(order *models.Order, login string) error {
	var loginFromDB string
	var timeFromDB time.Time
	timeCreated := time.Now()

	q := `INSERT INTO orders(number,user_id,status,accrual,uploaded_at)
	VALUES ($1,(SELECT id FROM users WHERE login=$2),$3,$4,$5)
	ON CONFLICT (number) DO UPDATE SET
	number=EXCLUDED.number
	RETURNING uploaded_at,(SELECT login FROM users WHERE id=orders.user_id);`

	row := l.db.QueryRow(context.Background(), q, order.Number, login, order.Status, order.Accrual, timeCreated)

	if err := row.Scan(&timeFromDB, &loginFromDB); err != nil {
		l.logger.Error(err)
		return ErrInt
	}
	if timeCreated.Format(time.StampMilli) != timeFromDB.Format(time.StampMilli) {
		if loginFromDB != login {
			return ErrOrdUsrConfl
		}
		return ErrOrdOverLap
	}

	return nil
}

//update order ========================================================
func (l *loyalty) UpdateOrder(order *models.Order) error {
	q := `UPDATE orders 
	SET status=$1,accrual=$2
		WHERE number=$3;`
	_, err := l.db.Exec(context.Background(), q, order.Status, order.Accrual, order.Number)
	if err != nil {
		l.logger.Error(err)
		return ErrInt
	}

	if order.Accrual > 0 {
		q = `UPDATE accounts 
		SET current=current+$1
			WHERE id=
		(SELECT user_id FROM orders
			WHERE number=$2);`
		_, err = l.db.Exec(context.Background(), q, order.Accrual, order.Number)
		if err != nil {
			l.logger.Error(err)
			return ErrInt
		}

		return nil
	}

	return nil
}

//get orders list ========================================================
func (l *loyalty) GetOrders(login string) ([]models.OrderDTO, error) {
	var accrual int
	q := `SELECT number, status, accrual, uploaded_at
	FROM orders
		WHERE
	user_id=(SELECT id FROM users WHERE login=$1);`

	rows, err := l.db.Query(context.Background(), q, login)
	if err != nil {
		l.logger.Error(err)
		return nil, ErrInt
	}
	var list = make([]models.OrderDTO, 0, 10)
	for rows.Next() {
		var order models.OrderDTO
		err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			l.logger.Error(err)
			return nil, ErrInt
		}
		order.UploadedAt.Format(time.RFC3339)
		order.Accrual = float64(accrual) / 100
		list = append(list, order)
	}
	return list, nil
}

//get customer balance ========================================================
func (l *loyalty) GetBalance(login string) (*models.Account, error) {
	q := `SELECT current, withdrawn
	FROM accounts 
		WHERE
	id=(SELECT account_id FROM users WHERE login=$1);`

	var account models.Account
	row := l.db.QueryRow(context.Background(), q, login)
	if err := row.Scan(&account.Current, &account.Withdrawn); err != nil {
		l.logger.Error(err)
		return nil, ErrInt
	}
	return &account, nil
}

//withdraw ========================================================
func (l *loyalty) Withdraw(withdraw *models.WithdrawalDTO, login string) error {
	tx, err := l.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		l.logger.Error(err)
		return ErrInt
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()
	var id int
	q := `INSERT INTO withdrawals
	 (order_id, sum, processed_at)
	 	VALUES ((
			 SELECT id FROM orders
		WHERE
			number=$1
		 ),$2,$3)
		 RETURNING id;`
	sum := int(withdraw.Sum * 100)
	res := l.db.QueryRow(context.Background(), q, withdraw.Order, sum, time.Now())

	if err := res.Scan(&id); err != nil {
		l.logger.Error(err)
		return ErrInt
	}

	q = `UPDATE accounts 
	SET current=current-$1,withdrawn=withdrawn+$1
		WHERE id=
	(SELECT account_id FROM users
		WHERE login=$2);`
	_, err = l.db.Exec(context.Background(), q, sum, login)
	if err != nil {
		l.logger.Error(err)
		return ErrInt
	}

	return nil
}

//get withdrawls ========================================================
func (l *loyalty) GetWithdrawls(login string) ([]models.WithdrawalDTO, error) {
	q := `SELECT number,sum,processed_at
	FROM withdrawals    
	JOIN 
		orders ON order_id=orders.id
	WHERE
		user_id=(SELECT id FROM users
	WHERE login=$1)
		ORDER BY processed_at;`

	rows, err := l.db.Query(context.Background(), q, login)
	if err != nil {
		l.logger.Error(err)
		return nil, ErrInt
	}
	var list = make([]models.WithdrawalDTO, 0, 10)
	for rows.Next() {
		var withdraw models.WithdrawalDTO
		var sum int
		err := rows.Scan(&withdraw.Order, &sum, &withdraw.ProcessedAt)
		withdraw.ProcessedAt.Format(time.RFC3339)
		withdraw.Sum = float64(sum / 100)
		if err != nil {
			l.logger.Error(err)
			return nil, ErrInt
		}
		list = append(list, withdraw)
	}
	return list, nil
}

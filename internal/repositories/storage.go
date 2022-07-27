package repositories

import (
	"context"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

type DBpgx struct {
	Conn *pgxpool.Pool
}

type Storage interface {
	CreateUser(user models.User) (uint, error)
	GetUser(user models.User) (uint, error)
	CheckOrder(number string) (uint, error)
	AddOrder(number string, userID uint) error
	GetListOrders(userid uint) ([]models.Order, error)
	BalanceUser(userid uint) (float32, error)
	SumWithdrawn(userid uint) (float32, error)
	ChangeBalance(change models.AccountBalance) error
	ListWithdraw(userid uint) ([]models.Withdraw, error)
}

func NewDB(pool *pgxpool.Pool) *DBpgx {
	return &DBpgx{Conn: pool}
}

func (d DBpgx) CreateUser(user models.User) (uint, error) {
	q := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id;`

	rows, err := d.Conn.Query(context.Background(), q, user.Username, user.Password)
	if err != nil {
		log.Print(err)
		log.Print("user not created")
		return 0, err
	}
	defer rows.Close()

	var id uint
	if rows.Next() {
		log.Print("rows next")
		err := rows.Scan(&id)
		if err != nil {
			log.Print(err)
			return 0, err
		}

	}
	log.Print(id)
	return id, nil
}

func (d DBpgx) GetUser(user models.User) (uint, error) {
	q := `SELECT id FROM users WHERE username = $1 and password = $2;`
	row, err := d.Conn.Query(context.Background(), q, user.Username, user.Password)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer row.Close()

	var id uint
	if row.Next() {
		err := row.Scan(&id)
		if err != nil {
			log.Print(err)
			return 0, err
		}
		log.Print("user found")
		return id, nil
	}
	log.Print("user no found")
	return 0, nil
}

func (d DBpgx) CheckOrder(number string) (uint, error) {
	q := `select userid from orders where ordernumber = $1`
	row, err := d.Conn.Query(context.Background(), q, number)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer row.Close()

	var id uint
	if row.Next() {
		err := row.Scan(&id)
		if err != nil {
			log.Print(err)
			return 0, err
		}
		log.Print("user found")
		return id, nil
	}
	log.Print("user no found")
	return 0, nil
}

func (d DBpgx) AddOrder(number string, userID uint) error {
	q := `INSERT INTO orders (ordernumber, date, userid, status) VALUES ($1, $2, $3, $4);`
	_, err := d.Conn.Exec(context.Background(), q, number, time.Now().Format(time.RFC3339), userID, "NEW")
	if err != nil {
		return err
	}
	return err
}

func (d DBpgx) GetListOrders(userid uint) ([]models.Order, error) {
	q := `select ordernumber, status, date from orders where userid=$1;`
	rows, err := d.Conn.Query(context.Background(), q, userid)
	if err != nil {
		return []models.Order{}, err
	}
	defer rows.Close()

	var ListOrders []models.Order
	for rows.Next() {
		x := models.Order{}
		err := rows.Scan(&x.OrderNumber, &x.Status, &x.Date)
		if err != nil {
			return []models.Order{}, err
		}
		ListOrders = append(ListOrders, x)
	}
	if rows.Err() != nil {
		return []models.Order{}, err
	}
	return ListOrders, err
}

func (d DBpgx) BalanceUser(userid uint) (float32, error) {
	q := `select id, balance from accountbalance where userid=$1 order by id DESC LIMIT 1;`
	rows, err := d.Conn.Query(context.Background(), q, userid)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer rows.Close()

	var balance float32
	if rows.Next() {
		x := 0
		err := rows.Scan(&x, &balance)
		if err != nil {
			log.Print(err)
			return 0, err
		}
	}
	log.Print(balance)
	return balance, err
}

func (d DBpgx) SumWithdrawn(userid uint) (float32, error) {
	q := `select userid, sum(sumaccrual) from accountbalance where userid = $1 and typemove = $2 group by userid;`
	rows, err := d.Conn.Query(context.Background(), q, userid, "withdraw")
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer rows.Close()

	var withdrawn float32
	if rows.Next() {
		x := 0
		err := rows.Scan(&x, &withdrawn)
		if err != nil {
			log.Print(err)
			return 0, err
		}
	}
	log.Print(withdrawn)

	return withdrawn, err
}

func (d DBpgx) ChangeBalance(change models.AccountBalance) error {
	q := `INSERT INTO accountbalance (userid, ordernumber, typemove, sumaccrual, date, balance) 
	VALUES  ($1, $2, $3, $4, $5 ,$6);`

	_, err := d.Conn.Exec(context.Background(), q, change.UserID, change.OrderNumber, change.TypeMove, change.SumAccrual, time.Now().Format(time.RFC3339), change.Balance)
	if err != nil {
		log.Print(err)
		return err
	}
	return err
}

func (d DBpgx) ListWithdraw(userid uint) ([]models.Withdraw, error) {
	q := `select ordernumber, sumaccrual, date from accountbalance where typemove = 'withdraw' and userid = $1;`

	rows, err := d.Conn.Query(context.Background(), q, userid)
	if err != nil {
		log.Print(err)
		return []models.Withdraw{}, err
	}
	defer rows.Close()

	var ListWithdraw []models.Withdraw
	for rows.Next() {
		x := models.Withdraw{}
		err := rows.Scan(&x.Order, &x.Sum, &x.ProcessedAt)
		if err != nil {
			return []models.Withdraw{}, err
		}
		ListWithdraw = append(ListWithdraw, x)
	}
	if rows.Err() != nil {
		return []models.Withdraw{}, err
	}
	return ListWithdraw, err
}

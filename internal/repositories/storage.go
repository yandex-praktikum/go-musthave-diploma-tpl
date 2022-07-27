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
	CheckOrder(number uint) (uint, error)
	AddOrder(number uint, userID uint) error
	GetListOrders(userid uint) ([]models.Order, error)
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

func (d DBpgx) CheckOrder(number uint) (uint, error) {
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

func (d DBpgx) AddOrder(number uint, userID uint) error {
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

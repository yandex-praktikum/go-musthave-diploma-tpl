package router

import (
	"context"
	"database/sql"
	"time"
)

var CreateTableOperations = `CREATE TABLE OperationsGopherMart(
order_number     	varchar(32),   
users          		varchar(64),
uploaded_at       	integer,
Status				varchar(32),
operation 			varchar(32),
points      		integer,
)`

var CreateTableUsers = `CREATE TABLE UsersGopherMart(
login				varchar(32),
users          		varchar(64),
current_points    	integer,
withdrawn_points  	integer,
)`

type Operation struct {
	Order_number string     `json:"number"`
	Status       string     `json:"Status"`
	Points       uint       `json:"accrual"`
	Uploaded_at  *time.Time `json:"uploaded_at"`
	Users        string     `json:"-"`
	//Operation    string     `json:"-"` // "accrual" or "withdraw"
}

const (
	accrual  = "accrual"
	withdraw = "withdraw"

	processing = "PROCESSING"
	registered = "REGISTERED"
	neworder   = "NEW"
	invalid    = "INVALID"
)

type DBI interface {
	Connect(connStr string) (err error)
	CreateTable() error
	Ping(ctx context.Context) error
	Close() error
}

type Database struct {
	connection *sql.DB
}

func InitDB() (*Database, error) {
	return &Database{}, nil
}

func (db *Database) Connect(connStr string) (err error) {
	db.connection, err = sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	if err = db.CreateTable(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err = db.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (db *Database) CreateTable() error {
	if _, err := db.connection.Exec("Drop TABLE OperationsGopherMart"); err != nil {
		return err
	}
	if _, err := db.connection.Exec("Drop TABLE UsersGopherMart"); err != nil {
		return err
	}
	if _, err := db.connection.Exec(CreateTableOperations); err != nil {
		return err
	}
	if _, err := db.connection.Exec("CREATE UNIQUE INDEX order_index ON OperationsGopherMart (order_number)"); err != nil {
		return err
	}
	if _, err := db.connection.Exec(CreateTableUsers); err != nil {
		return err
	}
	_, err := db.connection.Exec("CREATE UNIQUE INDEX login_index ON UsersGopherMart (login,users)")
	return err
}

func (db *Database) Ping(ctx context.Context) error {
	if err := db.connection.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (db *Database) Close() error {
	return db.connection.Close()
}

// отловить ошибку с уникальным номером заказа, там где будет вызываться
func (db *Database) WriteOrderAccrual(order string, user string) (err error) {
	_, err = db.connection.Exec("insert into OperationsGopherMart (order_number, users, operation) values ($1,$2,$3)", order, user, accrual)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ReadOrderAccrual(storage string, user string, order string) (op Operation, err error) {
	row := db.connection.QueryRow("select order_number, status, uploaded_at, points from OperationsGopherMart where order_number = $1 and users = $2 and operation != $3",
		order, user, accrual)
	err = row.Scan(&op.Order_number, &op.Status, &op.Uploaded_at, &op.Points)
	if err != nil {
		return Operation{}, err
	}
	if (op.Status == neworder) || (op.Status == processing) {
		op.Status, op.Points, err = db.UpdateOrderAccrual(storage, op.Order_number, user)
		if err != nil {
			return Operation{}, err
		}
	}
	return op, nil
}

// проверить ошибки, return
func (db *Database) ReadAllOrderAccrualUser(storage string, user string) (ops []Operation, err error) {
	var op Operation
	rows, err := db.connection.Query("select order_number, status, uploaded_at, points from OperationsGopherMart where users = $1 and operation != $2", user, accrual)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&op.Order_number, &op.Status, &op.Uploaded_at, &op.Points)
		if err != nil {
			return nil, err
		}
		if (op.Status == neworder) || (op.Status == processing) {
			op.Status, op.Points, err = db.UpdateOrderAccrual(storage, op.Order_number, user)
			if err != nil {
				return nil, err
			}
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func (db *Database) UpdateOrderAccrual(storage string, order string, user string) (status string, points uint, err error) {
	orderStatus, err := accrualOrderStatus(storage, order)
	if err != nil {
		return "", 0, err
	}
	_, err = db.connection.Exec("UPDATE OperationsGopherMart SET status = $1,point = $2 WHERE order=$3", orderStatus.Status, orderStatus.Point, order)
	if err != nil {
		return "", 0, err
	}
	//зачислить балы пользователю
	if status == registered {
		_, err = db.connection.Exec("UPDATE UsersGopherMart SET current_points = current_points + $1 WHERE users=$2", orderStatus.Point, user)
		if err != nil {
			return "", 0, err
		}
	}
	return orderStatus.Status, orderStatus.Point, nil
}

type UserPoints struct {
	CurrentPoints   int `json:"current"`
	WithdrawnPoints int `json:"withdrawn"`
}

// информация по счёту, но перед этим ReadAllOrderAccrualUser
func (db *Database) ReadUserPoints(storage string, user string) (u UserPoints, err error) {
	if _, err = db.ReadAllOrderAccrualUser(storage, user); err != nil {
		return UserPoints{}, err
	}
	row := db.connection.QueryRow("select current_points, withdrawn_points from UsersGopherMart where users = $1",
		user)
	if err = row.Scan(&u.CurrentPoints, &u.WithdrawnPoints); err != nil {
		return UserPoints{}, err
	}
	return u, nil
}

// списание
// попытка списать баллы, но перед этим ReadAllOrderAccrualUser
func (db *Database) WithdrawnUserPoints(storage string, user string, sum int) (err error) {
	var u UserPoints
	//
	if _, err = db.ReadAllOrderAccrualUser(storage, user); err != nil {
		return err
	}
	u, err = db.ReadUserPoints(storage, user)
	if err != nil {
		return err
	}
	if u.CurrentPoints >= sum {
		// списываем и ретурн
		_, err = db.connection.Exec("UPDATE UsersGopherMart SET current_points = current_points - $1 and withdrawn_points = withdrawn_points + $1 WHERE users=$2",
			sum, user)
		if err != nil {
			return err
		}

	}

	// ошибка, недостаточно баллов и ретурн
	return nil
}

// создание заказа со списанием

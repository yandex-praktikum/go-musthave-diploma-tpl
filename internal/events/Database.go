package events

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"GopherMart/internal/errorsgm"
)

var createTableOperations = `CREATE TABLE OperationsGopherMart(
order_number     	varchar(32),   
login          		varchar(32),
uploaded_at       	varchar(32),
status				varchar(32),
operation 			varchar(32),
points      		integer
)`

var createTableUsers = `CREATE TABLE UsersGopherMart(
login				varchar(32),
password          	varchar(32),
current_points    	integer,
withdrawn_points  	integer
)`

type Operation struct {
	OrderNumber string  `json:"number"`
	Status      string  `json:"status"`
	Points      float64 `json:"accrual,omitempty"`
	UploadedAt  string  `json:"uploaded_at"`
}

type OperationO struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"-"`
	Points      float64 `json:"accrual,omitempty"`
	UploadedAt  string  `json:"uploaded_at"`
}

const (
	accrual  = "accrual"
	withdraw = "withdraw"

	newOrder   = "NEW"
	processing = "PROCESSING"
	processed  = "PROCESSED"
	//invalid    = "INVALID"
)

type DBI interface {
	Connect(connStr string) (err error)
	CreateTable() error
	Ping(ctx context.Context) error
	Close() error

	RegisterUser(login string, pass string) (tokenJWT string, err error)
	LoginUser(login string, pass string) (tokenJWT string, err error)

	WriteOrderAccrual(order string, user string) (err error)
	ReadAllOrderAccrualUser(user string) (ops []Operation, err error)
	ReadUserPoints(user string) (u UserPoints, err error)
	WithdrawnUserPoints(user string, order string, sum float64) (err error)
	WriteOrderWithdrawn(order string, user string, point float64) (err error)
	ReadAllOrderWithdrawnUser(user string) (ops []Operation, err error)

	ReadAllOrderAccrualNoComplite() (orders []orderstruct, err error)
	UpdateOrderAccrual(login string, orderAccrual requestAccrual) (err error)
}

type Database struct {
	connection *sql.DB
}

func InitDB() (*Database, error) {
	return &Database{}, nil
}

func (db *Database) Connect(connStr string) (err error) {

	db.connection, err = sql.Open("pgx", connStr)
	//db.connection, err = sql.Open("pgx", "postgres://postgres:0000@localhost:5432/postgres")    //
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
	//db.connection.Exec("Drop TABLE OperationsGopherMart")
	//db.connection.Exec("Drop TABLE UsersGopherMart")

	if _, err := db.connection.Exec(createTableOperations); err != nil {
		return err
	}
	_, err := db.connection.Exec("CREATE UNIQUE INDEX order_index ON OperationsGopherMart (order_number)")
	if err != nil {
		return err
	}

	if _, err = db.connection.Exec(createTableUsers); err != nil {
		return err
	}
	if _, err = db.connection.Exec("CREATE UNIQUE INDEX login_index ON UsersGopherMart (login)"); err != nil {
		return err
	}
	return nil
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

// добавление заказа для начисления
func (db *Database) WriteOrderAccrual(order string, user string) (err error) {
	timeNow := time.Now().Format(time.RFC3339)

	var loginOrder string

	rows, err := db.connection.Query("select login from OperationsGopherMart where order_number = $1", order)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&loginOrder)
		if err != nil {
			return err
		}
	}

	if loginOrder != "" {
		if loginOrder == user {
			return errorsgm.ErrLoadedEarlierThisUser // надо что то вернуть
		}
		return errorsgm.ErrLoadedEarlierAnotherUser
	}
	_, err = db.connection.Exec("insert into OperationsGopherMart (order_number, login, operation, uploaded_at, status, points) values ($1,$2,$3,$4,$5,$6)",
		order, user, accrual, timeNow, newOrder, 0)
	if err != nil {
		return err
	}
	return nil
}

// вывод всех заказов пользователя
func (db *Database) ReadAllOrderAccrualUser(user string) (ops []Operation, err error) {
	var op Operation
	rows, err := db.connection.Query("select order_number, status, uploaded_at, points  from OperationsGopherMart where login = $1 ORDER BY uploaded_at ASC", user)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&op.OrderNumber, &op.Status, &op.UploadedAt, &op.Points)
		if err != nil {
			return nil, err
		}
		op.Points = op.Points / 100
		ops = append(ops, op)
	}

	return ops, nil
}

type UserPoints struct {
	CurrentPoints   float64 `json:"current"`
	WithdrawnPoints float64 `json:"withdrawn"`
}

// информация о потраченных и остатках баллов
func (db *Database) ReadUserPoints(user string) (up UserPoints, err error) {
	row := db.connection.QueryRow("select current_points, withdrawn_points from UsersGopherMart where login = $1",
		user)
	if err = row.Scan(&up.CurrentPoints, &up.WithdrawnPoints); err != nil {
		return UserPoints{}, err
	}
	up.WithdrawnPoints = up.WithdrawnPoints / 100
	up.CurrentPoints = up.CurrentPoints / 100
	return up, nil
}

// списание
func (db *Database) WithdrawnUserPoints(user string, order string, sum float64) (err error) {
	var u UserPoints

	u, err = db.ReadUserPoints(user)
	if err != nil {
		return err
	}
	if u.CurrentPoints < sum {
		return errorsgm.ErrDontHavePoints
	}

	err = db.WriteOrderWithdrawn(user, order, sum)
	if err != nil {
		return err
	}

	_, err = db.connection.Exec("UPDATE UsersGopherMart SET current_points = current_points - $1,withdrawn_points = withdrawn_points + $2 WHERE login=$3",
		sum*100, sum*100, user)
	if err != nil {
		fmt.Println("===postAPIUserBalanceWithdraw=6=", err)
		return err
	}

	return nil
}

func (db *Database) WriteOrderWithdrawn(user string, order string, point float64) (err error) {
	timeNow := time.Now().Format(time.RFC3339)

	_, err = db.connection.Exec("insert into OperationsGopherMart (order_number, login, operation, uploaded_at, status,  points) values ($1,$2,$3,$4,$5,$6)",
		order, user, withdraw, timeNow, processed, point*100)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ReadAllOrderWithdrawnUser(user string) (ops []OperationO, err error) {
	var op OperationO
	rows, err := db.connection.Query("select order_number, status, uploaded_at, points from OperationsGopherMart where login = $1 and operation != $2", user, withdraw)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&op.OrderNumber, &op.Status, &op.UploadedAt, &op.Points)
		if err != nil {
			return nil, err
		}
		op.Points = op.Points / 100
		ops = append(ops, op)
	}
	return ops, nil
}

// регистрация
func (db *Database) RegisterUser(login string, pass string) (tokenJWT string, err error) {
	h := md5.New()
	h.Write([]byte(pass))
	passHex := hex.EncodeToString(h.Sum(nil))

	_, err = db.connection.Exec("insert into UsersGopherMart (login, password, current_points, withdrawn_points ) values ($1,$2,$3,$4)", login, passHex, 0, 0)
	if err != nil {
		return "", err
	}

	tokenJWT, err = EncodeJWT(login)
	if err != nil {
		return "", err
	}
	return tokenJWT, nil
}

// авторизация
func (db *Database) LoginUser(login string, pass string) (tokenJWT string, err error) {
	h := md5.New()
	h.Write([]byte(pass))
	pass = hex.EncodeToString(h.Sum(nil))
	var dbPass string

	row := db.connection.QueryRow("select password from UsersGopherMart where login = $1",
		login)
	if err = row.Scan(&dbPass); err != nil {
		return "", err
	}
	if dbPass != pass {
		return "", nil
	}
	tokenJWT, err = EncodeJWT(login)
	if err != nil {
		return "", err
	}
	return tokenJWT, nil
}

type orderstruct struct {
	Order string
	Login string
}

func (db *Database) ReadAllOrderAccrualNoComplite() (orders []orderstruct, err error) {
	var order orderstruct
	rows, err := db.connection.Query("select order_number,login from OperationsGopherMart where status = $1 or status = $2",
		newOrder, processing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&order.Order, &order.Login)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (db *Database) UpdateOrderAccrual(login string, orderAccrual requestAccrual) (err error) {
	_, err = db.connection.Exec("UPDATE OperationsGopherMart SET status = $1,points = $2 WHERE order_number=$3",
		orderAccrual.Status, orderAccrual.Accrual, orderAccrual.Order)
	if err != nil {
		return err
	}
	//зачислить балы пользователю
	if orderAccrual.Status == processed {
		_, err = db.connection.Exec("UPDATE UsersGopherMart SET current_points = current_points + $1 WHERE login=$2",
			orderAccrual.Accrual, login)
		if err != nil {
			return err
		}
	}
	return nil
}

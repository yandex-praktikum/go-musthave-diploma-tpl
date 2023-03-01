package events

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"GopherMart/internal/errorsgm"
)

var createTableOperations = `CREATE TABLE operations_gopher_mart(
order_number     	varchar(32),   
login          		varchar(32),
uploaded_at       	varchar(32),
status				varchar(32),
operation 			varchar(32),
points      		integer
)`

var createTableUsers = `CREATE TABLE users_gopher_mart(
login				varchar(32),
password          	varchar(32),
current_points    	integer,
withdrawn_points  	integer
)`

type OperationAccrual struct {
	OrderNumber string  `json:"number"`
	Status      string  `json:"status"`
	Points      float64 `json:"accrual,omitempty"`
	UploadedAt  string  `json:"uploaded_at"`
	Operation   string  `json:"-"`
}

type OperationWithdraw struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"-"`
	Points      float64 `json:"sum,omitempty"`
	UploadedAt  string  `json:"uploaded_at"`
	Operation   string  `json:"-"`
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
	ReadAllOrderAccrualUser(user string) (ops []OperationAccrual, err error)
	ReadUserPoints(user string) (u UserPoints, err error)
	WithdrawnUserPoints(user string, order string, sum float64) (err error)
	WriteOrderWithdrawn(order string, user string, point float64) (err error)
	ReadAllOrderWithdrawnUser(user string) (ops []OperationWithdraw, err error)

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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//db.connection.Exec("Drop TABLE operations_gopher_mart")
	//db.connection.Exec("Drop TABLE users_gopher_mart")

	if _, err := db.connection.ExecContext(ctx, createTableOperations); err != nil {
		return err
	}
	_, err := db.connection.ExecContext(ctx, "CREATE UNIQUE INDEX order_index ON operations_gopher_mart (order_number)")
	if err != nil {
		return err
	}

	if _, err = db.connection.ExecContext(ctx, createTableUsers); err != nil {
		return err
	}
	if _, err = db.connection.ExecContext(ctx, "CREATE UNIQUE INDEX login_index ON users_gopher_mart (login)"); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	timeNow := time.Now().Format(time.RFC3339)
	var loginOrder string

	rows, err := db.connection.QueryContext(ctx, "select login from operations_gopher_mart where order_number = $1", order)
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
	err = rows.Err()
	if err != nil {
		return err
	}

	if loginOrder != "" {
		if loginOrder == user {
			return errorsgm.ErrLoadedEarlierThisUser // надо что то вернуть
		}
		return errorsgm.ErrLoadedEarlierAnotherUser
	}
	_, err = db.connection.ExecContext(ctx, "insert into operations_gopher_mart (order_number, login, operation, uploaded_at, status, points) values ($1,$2,$3,$4,$5,$6)",
		order, user, accrual, timeNow, newOrder, 0)
	if err != nil {
		return err
	}
	return nil
}

// вывод всех заказов пользователя
func (db *Database) ReadAllOrderAccrualUser(user string) (ops []OperationAccrual, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var op OperationAccrual
	rows, err := db.connection.QueryContext(ctx, "select order_number, status, uploaded_at, points, operation from operations_gopher_mart where login = $1 ORDER BY uploaded_at ASC", user)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&op.OrderNumber, &op.Status, &op.UploadedAt, &op.Points, &op.Operation)
		if err != nil {
			return nil, err
		}
		if op.Operation == accrual {
			op.Points /= 100
			ops = append(ops, op)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ops, nil
}

type UserPoints struct {
	CurrentPoints   float64 `json:"current"`
	WithdrawnPoints float64 `json:"withdrawn"`
}

// информация о потраченных и остатках баллов
func (db *Database) ReadUserPoints(user string) (up UserPoints, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	row := db.connection.QueryRowContext(ctx, "select current_points, withdrawn_points from users_gopher_mart where login = $1",
		user)
	if err = row.Scan(&up.CurrentPoints, &up.WithdrawnPoints); err != nil {
		return UserPoints{}, err
	}
	up.WithdrawnPoints /= 100
	up.CurrentPoints /= 100
	return up, nil
}

// списание
func (db *Database) WithdrawnUserPoints(user string, order string, sum float64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
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

	_, err = db.connection.ExecContext(ctx, "UPDATE users_gopher_mart SET current_points = current_points - $1,withdrawn_points = withdrawn_points + $2 WHERE login=$3",
		sum*100, sum*100, user)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) WriteOrderWithdrawn(user string, order string, point float64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	timeNow := time.Now().Format(time.RFC3339)

	_, err = db.connection.ExecContext(ctx, "insert into operations_gopher_mart (order_number, login, operation, uploaded_at, status,  points) values ($1,$2,$3,$4,$5,$6)",
		order, user, withdraw, timeNow, processed, point*100)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ReadAllOrderWithdrawnUser(user string) (ops []OperationWithdraw, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var op OperationWithdraw
	rows, err := db.connection.QueryContext(ctx, "select order_number, status, uploaded_at, points, operation from operations_gopher_mart where login = $1", user)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&op.OrderNumber, &op.Status, &op.UploadedAt, &op.Points, &op.Operation)
		if err != nil {
			return nil, err
		}
		if op.Operation == withdraw {
			op.Points /= 100
			ops = append(ops, op)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ops, nil
}

// регистрация
func (db *Database) RegisterUser(login string, pass string) (tokenJWT string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	h := md5.New()
	h.Write([]byte(pass))
	passHex := hex.EncodeToString(h.Sum(nil))

	_, err = db.connection.ExecContext(ctx, "insert into users_gopher_mart (login, password, current_points, withdrawn_points ) values ($1,$2,$3,$4)", login, passHex, 0, 0)
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	h := md5.New()
	h.Write([]byte(pass))
	pass = hex.EncodeToString(h.Sum(nil))
	var dbPass string

	row := db.connection.QueryRowContext(ctx, "select password from users_gopher_mart where login = $1",
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var order orderstruct
	rows, err := db.connection.QueryContext(ctx, "select order_number,login from operations_gopher_mart where status = $1 or status = $2",
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
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (db *Database) UpdateOrderAccrual(login string, orderAccrual requestAccrual) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = db.connection.ExecContext(ctx, "UPDATE operations_gopher_mart SET status = $1,points = $2 WHERE order_number=$3",
		orderAccrual.Status, orderAccrual.Accrual, orderAccrual.Order)
	if err != nil {
		return err
	}
	//зачислить балы пользователю
	if orderAccrual.Status == processed {
		_, err = db.connection.ExecContext(ctx, "UPDATE users_gopher_mart SET current_points = current_points + $1 WHERE login=$2",
			orderAccrual.Accrual, login)
		if err != nil {
			return err
		}
	}
	return nil
}

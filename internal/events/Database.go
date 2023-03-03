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
	Connect(ctx context.Context, connStr string) (err error)
	CreateTable(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error

	RegisterUser(ctx context.Context, login string, pass string) (tokenJWT string, err error)
	LoginUser(ctx context.Context, login string, pass string) (tokenJWT string, err error)

	WriteOrderAccrual(ctx context.Context, order string, user string) (err error)
	ReadAllOrderAccrualUser(ctx context.Context, user string) (ops []OperationAccrual, err error)
	ReadUserPoints(ctx context.Context, user string) (u UserPoints, err error)
	WithdrawnUserPoints(ctx context.Context, user string, order string, sum float64) (err error)
	WriteOrderWithdrawn(ctx context.Context, order string, user string, point float64) (err error)
	ReadAllOrderWithdrawnUser(ctx context.Context, user string) (ops []OperationWithdraw, err error)

	ReadAllOrderAccrualNoComplite(ctx context.Context) (orders []orderstruct, err error)
	UpdateOrderAccrual(ctx context.Context, login string, orderAccrual requestAccrual) (err error)
}

type Database struct {
	connection *sql.DB
}

func InitDB() (*Database, error) {
	return &Database{}, nil
}

func (db *Database) Connect(ctx context.Context, connStr string) (err error) {
	db.connection, err = sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	if err = db.CreateTable(ctx); err != nil {
		return err
	}
	if err = db.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (db *Database) CreateTable(ctx context.Context) error {
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
func (db *Database) WriteOrderAccrual(ctx context.Context, order string, user string) (err error) {
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
func (db *Database) ReadAllOrderAccrualUser(ctx context.Context, user string) (ops []OperationAccrual, err error) {
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
func (db *Database) ReadUserPoints(ctx context.Context, user string) (up UserPoints, err error) {
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
func (db *Database) WithdrawnUserPoints(ctx context.Context, user string, order string, sum float64) (err error) {
	var u UserPoints

	u, err = db.ReadUserPoints(ctx, user)
	if err != nil {
		return err
	}
	if u.CurrentPoints < sum {
		return errorsgm.ErrDontHavePoints
	}

	err = db.WriteOrderWithdrawn(ctx, user, order, sum)
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

func (db *Database) WriteOrderWithdrawn(ctx context.Context, user string, order string, point float64) (err error) {
	timeNow := time.Now().Format(time.RFC3339)

	_, err = db.connection.ExecContext(ctx, "insert into operations_gopher_mart (order_number, login, operation, uploaded_at, status,  points) values ($1,$2,$3,$4,$5,$6)",
		order, user, withdraw, timeNow, processed, point*100)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ReadAllOrderWithdrawnUser(ctx context.Context, user string) (ops []OperationWithdraw, err error) {
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
func (db *Database) RegisterUser(ctx context.Context, login string, pass string) (tokenJWT string, err error) {

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
func (db *Database) LoginUser(ctx context.Context, login string, pass string) (tokenJWT string, err error) {

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

func (db *Database) ReadAllOrderAccrualNoComplite(ctx context.Context) (orders []orderstruct, err error) {

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

func (db *Database) UpdateOrderAccrual(ctx context.Context, login string, orderAccrual requestAccrual) (err error) {

	_, err = db.connection.ExecContext(ctx, "UPDATE operations_gopher_mart SET status = $1,points = $2 WHERE order_number=$3",
		orderAccrual.Status, orderAccrual.Accrual, orderAccrual.Order)
	if err != nil {
		return err
	}

	if orderAccrual.Status == processed {
		_, err = db.connection.ExecContext(ctx, "UPDATE users_gopher_mart SET current_points = current_points + $1 WHERE login=$2",
			orderAccrual.Accrual, login)
		if err != nil {
			return err
		}
	}
	return nil
}

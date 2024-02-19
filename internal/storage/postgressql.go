package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Azcarot/GopherMarketProject/internal/utils"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const SecretKey string = "super-secret"

type MyCustomClaims struct {
	jwt.MapClaims
}

type UserData struct {
	Login       string
	Password    string
	BonusPoints int
	Date        string
}
type OrderData struct {
	OrderNumber uint64 `json:"number"`
	Accural     int    `json:"accrual"`
	User        string
	State       string `json:"status"`
	Date        string `json:"uploaded_at"`
}

type OrderResponse struct {
	OrderNumber string `json:"number"`
	Accural     int    `json:"accrual"`
	State       string `json:"status"`
	Date        string `json:"uploaded_at"`
}

var DB *pgx.Conn

type pgxConnTime struct {
	attempts          int
	timeBeforeAttempt int
}

func NewConn(f utils.Flags) error {
	var err error
	var attempts pgxConnTime
	attempts.attempts = 3
	attempts.timeBeforeAttempt = 1
	err = connectToDB(f)
	for err != nil {
		//если ошибка связи с бд, то это не эскпортируемый тип, отличный от PgError
		var pqErr *pgconn.PgError
		if errors.Is(err, pqErr) {
			return err

		}
		if attempts.attempts == 0 {
			return err
		}
		times := time.Duration(attempts.timeBeforeAttempt)
		time.Sleep(times * time.Second)
		attempts.attempts -= 1
		attempts.timeBeforeAttempt += 2
		err = connectToDB(f)

	}
	return nil
}

func connectToDB(f utils.Flags) error {
	var err error
	ps := fmt.Sprintf(f.FlagDBAddr)
	DB, err = pgx.Connect(context.Background(), ps)
	return err
}

func CheckDBConnection(db *pgx.Conn) http.Handler {
	checkConnection := func(res http.ResponseWriter, req *http.Request) {

		err := DB.Ping(context.Background())
		result := (err == nil)
		if result {
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusInternalServerError)
		}

	}
	return http.HandlerFunc(checkConnection)
}

func CreateTablesForGopherStore(db *pgx.Conn) {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL NOT NULL PRIMARY KEY, 
		login text NOT NULL, 
		password text NOT NULL, 
		accural_points bigint, 
		created text )`
	ctx := context.Background()

	_, err := db.Exec(ctx, query)

	if err != nil {

		log.Printf("Error %s when creating user table", err)

	}
	queryForFun := `DROP TABLE IF EXISTS orders CASCADE`
	db.Exec(ctx, queryForFun)
	query = `CREATE TABLE IF NOT EXISTS orders(
		id SERIAL NOT NULL PRIMARY KEY,
		order_number BIGINT,
		accural_points BIGINT,
		state TEXT,
		customer TEXT NOT NULL,
		created TEXT
	)`
	_, err = db.Exec(ctx, query)

	if err != nil {

		log.Printf("Error %s when creating order table", err)

	}
}

func CreateNewUser(db *pgx.Conn, data UserData) error {
	ctx := context.Background()
	encodedPW := utils.ShaData(data.Password, SecretKey)
	_, err := db.Exec(ctx, `INSERT into users (login, password, created) values ($1, $2, $3);`, data.Login, encodedPW, data.Date)
	return err
}

func CheckUserExists(db *pgx.Conn, data UserData) (bool, error) {
	ctx := context.Background()
	var login string
	sqlQuery := fmt.Sprintf(`SELECT login FROM users WHERE login = '%s'`, data.Login)
	err := db.QueryRow(ctx, sqlQuery).Scan(&login)

	if err == pgx.ErrNoRows {

		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil

}

func CheckUserPassword(db *pgx.Conn, data UserData) (bool, error) {
	encodedPw := utils.ShaData(data.Password, SecretKey)
	ctx := context.Background()
	sqlQuery := fmt.Sprintf(`SELECT login, password FROM users WHERE login = '%s'`, data.Login)
	var login, pw string
	err := db.QueryRow(ctx, sqlQuery).Scan(&login, &pw)
	if err != nil {
		return false, err
	}

	if encodedPw != pw {
		return false, nil
	}
	return true, nil
}

func CreateNewOrder(db *pgx.Conn, data OrderData) error {
	ctx := context.Background()
	data.State = "NEW"
	_, err := db.Exec(ctx, `insert into orders (order_number, accural_points, state, customer, created) values ($1, $2, $3, $4, $5);`, data.OrderNumber, data.Accural, data.State, data.User, data.Date)
	return err
}

func GetBalance(db *pgx.Conn, data OrderData) (int, int, error) {
	ctx := context.Background()
	rows, err := db.Query(ctx, "SELECT SUM(accural_points), SUM(withdrawal) FROM orders WHERE order_number = ?", data.OrderNumber)
	if err != nil {
		return 0, 0, err
	}

	defer rows.Close()
	var totalReward, withdrawn int
	for rows.Next() {
		if err := rows.Scan(&totalReward, &withdrawn); err != nil {
			return 0, 0, err
		}

	}

	return totalReward, withdrawn, nil
}

func VerifyToken(token string) (jwt.MapClaims, bool) {
	hmacSecretString := SecretKey
	hmacSecret := []byte(hmacSecretString)
	gettoken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return nil, false
	}

	if claims, ok := gettoken.Claims.(jwt.MapClaims); ok && gettoken.Valid {
		return claims, true

	} else {
		log.Printf("Invalid JWT Token")
		return nil, false
	}
}

func GetCustomerOrders(db *pgx.Conn, login string) ([]OrderResponse, error) {
	query := fmt.Sprintf(`SELECT order_number, accural_points, state, created FROM orders WHERE customer = '%s' ORDER BY id DESC`, login)
	result := []OrderResponse{}
	ctx := context.Background()
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var order OrderResponse
		if err := rows.Scan(&order.OrderNumber, &order.Accural, &order.State, &order.Date); err != nil {
			return result, err
		}
		result = append(result, order)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	return result, nil

}

func CheckIfOrderExists(db *pgx.Conn, data OrderData) (bool, bool) {
	query := fmt.Sprintf(`SELECT order_number, customer FROM orders WHERE order_number = %d`, data.OrderNumber)
	ctx := context.Background()
	var number uint64
	var login string
	err := db.QueryRow(ctx, query).Scan(&number, &login)
	if err == pgx.ErrNoRows {
		//No order
		return true, false
	}
	// Order exists for another user
	if login != data.User {
		return false, true
	}
	// order already exists for current user
	return false, false
}

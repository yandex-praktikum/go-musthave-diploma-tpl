package server

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", *PsqlInfo)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	_, err = DB.Exec(
		"CREATE TABLE IF NOT EXISTS users (" +
			"login VARCHAR (50) UNIQUE NOT NULL," +
			"password VARCHAR (50) NOT NULL" +
			")",
	)
	prevErr := errors.Join(nil, err)
	_, err = DB.Exec(
		"CREATE TABLE IF NOT EXISTS tokens (" +
			"token VARCHAR (1000) UNIQUE NOT NULL," +
			"name VARCHAR (50) UNIQUE NOT NULL," +
			"FOREIGN KEY (name) REFERENCES users(login)," +
			"expired_time TIMESTAMP NOT NULL" +
			")",
	)
	prevErr = errors.Join(prevErr, err)
	_, err = DB.Exec(
		"CREATE TABLE IF NOT EXISTS orders (" +
			"order_id INTEGER UNIQUE NOT NULL," +
			"name VARCHAR (50) NOT NULL," +
			"created_time TIMESTAMP NOT NULL," +
			"FOREIGN KEY (name) REFERENCES users(login)" +
			")",
	)
	prevErr = errors.Join(prevErr, err)
	return prevErr
}

type RegisterData struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type Token struct {
	Token       string `json:"token"`
	ExpiredDate int64  `json:"expired_time"`
}

type OrderInfo struct {
	Status      string    `json:"status"`
	CreatedTime time.Time `json:"uploaded_at"`
	Name        string    `json:"name"`
	OrderID     int64     `json:"number"`
	Accrual     int       `json:"accrual,omitempty"`
}

type Orders struct {
	Values []OrderInfo `json:"orders"`
}

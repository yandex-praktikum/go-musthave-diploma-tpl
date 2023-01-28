package main

import (
	"context"
	"database/sql"
	"time"
)

var СreateTableOperations = `CREATE TABLE OperationsGopherMart(
users          		varchar(64),
order_number     	varchar(32),
uploaded_at       	varchar(32),
status				varchar(32),
rewards_points      varchar(32),
)`

var СreateTableUsers = `CREATE TABLE UsersGopherMart(
login				varchar(32),
users          		varchar(64),
balance_points    	varchar(32),
withdrawn_points  	varchar(32),
)`

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
	if _, err := db.connection.Exec(СreateTableOperations); err != nil {
		return err
	}
	if _, err := db.connection.Exec("CREATE UNIQUE INDEX order_index ON OperationsGopherMart (order_number)"); err != nil {
		return err
	}
	if _, err := db.connection.Exec(СreateTableUsers); err != nil {
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

package db

import (
	"database/sql"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type StorageUsers interface {
	SaveTableUser(login, passwordHash string) error
	SaveTableUserAndUpdateToken(login, accessToken string) error
	Close() error
}

type DateBase struct {
	storage *sql.DB
}

func NewDB(logs *logger.Logger, addressConDB string) (*DateBase, error) {
	pstgr := &DateBase{}
	err := pstgr.initDB(logs, addressConDB)
	return pstgr, err
}

// InitDB инициализирует подключение к базе данных и создаем базу
func (d *DateBase) initDB(logs *logger.Logger, dataSourceName string) error {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return err
	}
	d.storage = db
	fmt.Println(dataSourceName)
	fmt.Println(db)

	err = d.createTableIfNotExists()
	if err != nil {
		return err
	}
	return nil
}

func (d *DateBase) createTableIfNotExists() error {
	// Создание таблицы users
	queryUsers := `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            login TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            access_token TEXT
        );
    `
	_, err := d.storage.Exec(queryUsers)
	if err != nil {
		return err
	}

	// Создание таблицы loyalty
	queryLoyalty := `
        CREATE TABLE IF NOT EXISTS loyalty (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL,
            order_id TEXT NOT NULL,
            bonus INT,
            order_status TEXT,
            is_deleted BOOL NOT NULL DEFAULT FALSE,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `
	_, err = d.storage.Exec(queryLoyalty)
	if err != nil {
		return err
	}

	// добавляем столбец если его ранее не было
	queryUpdateDataOrder := `
		ALTER TABLE loyalty ADD COLUMN IF NOT EXISTS created_in TIMESTAMP WITH TIME ZONE
`
	_, err = d.storage.Exec(queryUpdateDataOrder)
	if err != nil {
		return err
	}
	return nil
}

func (d *DateBase) Close() error {
	return d.storage.Close()
}

package postgres

import (
	"database/sql"
	"log"
	"time"
)

type PostgresDB struct {
	db *sql.DB
}

func New(dsn string) (*PostgresDB, error) {
	var counts int
	var connection *sql.DB
	var err error
	for {
		connection, err = openDB(dsn)
		if err != nil {
			log.Println("Database not ready...")
			counts++
		} else {
			log.Println("Connected to database")
			break
		}
		if counts > 2 {
			return nil, err
		}
		log.Printf("Retrying to connect after %d seconds\n", counts+2)
		time.Sleep(time.Duration(2+counts) * time.Second)
	}

	// TODO: DB INIT MIGRATION

	return &PostgresDB{
		db: connection,
	}, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

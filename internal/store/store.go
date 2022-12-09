package store

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func New() *Store {
	return &Store{}
}

func (s *Store) Open(databaseURI string) error {
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

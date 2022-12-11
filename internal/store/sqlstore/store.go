package sqlstore

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Store struct {
	db              *sql.DB
	userRepository  *UserRepository
	orderRepository *OrderRepository
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

func (s *Store) CreateTables() error {
	_, err := s.db.Exec(
		"CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, login VARCHAR(255) NOT NULL UNIQUE, encrypted_password VARCHAR(255) NOT NULL)",
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = s.db.Exec(
		"CREATE TABLE IF NOT EXISTS orders (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL, status VARCHAR(255) NOT NULL , number VARCHAR(255) NOT NULL UNIQUE, accrual FLOAT, uploaded_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL)",
	)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) User() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}
	return s.userRepository
}

func (s *Store) Order() *OrderRepository {
	if s.orderRepository != nil {
		return s.orderRepository
	}

	s.orderRepository = &OrderRepository{
		store: s,
	}
	return s.orderRepository
}

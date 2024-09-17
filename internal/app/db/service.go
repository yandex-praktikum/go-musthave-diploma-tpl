package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPgDB(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetConnMaxIdleTime(cfg.MaxLifetimeConn)

	return db, nil
}

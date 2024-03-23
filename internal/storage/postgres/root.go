package postgres

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/config"
)

type Storage struct {
	DB *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logrus.WithError(err).Fatal("failed to make DB connection")
		return nil, err
	}

	strg := &Storage{
		DB: db,
	}
	return strg, nil
}

func (s Storage) Migrate(cfg *config.Config, cmd string) error {
	if err := goose.Run(cmd, s.DB, cfg.MigrationsPath); err != nil {
		logrus.WithError(err).Fatal("migration error")
		return err
	}
	return nil
}

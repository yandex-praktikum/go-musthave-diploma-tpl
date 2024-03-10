package postgres

import (
	"database/sql"
	"fmt"
	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/pressly/goose"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	conn *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection db: %w", err)
	}

	storage := &Storage{conn: conn}

	logrus.Info("DB conn was initiated")
	return storage, nil
}

func (s Storage) MigrateUP(cfg *config.Config) error {
	if err := goose.Run("up", s.conn, cfg.MigrationsPath); err != nil {
		logrus.WithError(err).Fatal("migration error")
		return err
	}
	return nil
}

package migrate

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Cmd struct {
	m *migrate.Migrate
}

func NewMigrateCmd(cfg *Config) (*Cmd, error) {
	m, err := migrate.New(
		"file://migrations",
		cfg.Dsn,
	)
	if err != nil {
		return nil, err
	}

	return &Cmd{m: m}, nil
}

func (c *Cmd) Up() error {
	if err := c.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (c *Cmd) Down() error {
	if err := c.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewDatabase(ctx context.Context, dsn string, maxAttempts string) (db *Repository, err error) {
	var pool *pgxpool.Pool

	repetition, err := strconv.Atoi(maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("convert countRepetitionBD err: %w", err)
	}

	err = doWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, dsn)
		if err != nil {
			return err
		}

		return nil
	}, repetition, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	ok, err := startMigration(dsn)
	if err != nil && !ok {
		return nil, fmt.Errorf("failed migrate database: %w", err)
	}

	return &Repository{db: pool}, err
}

func startMigration(dsn string) (bool, error) {
	m, err := migrate.New("file://internal/database/postgres/migrations", dsn)
	if err != nil {
		if err != migrate.ErrNoChange {
			return false, err
		}
	}

	if err = m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return false, err
		}
	}
	return true, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

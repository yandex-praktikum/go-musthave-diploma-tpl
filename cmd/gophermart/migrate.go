package main

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	pgx5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/nikitaw13/go-musthave-diploma/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// runMigrations applies the embedded migrations to the database.
// It is a no-op when the schema is already up to date.
func runMigrations(pool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	dbDriver, err := pgx5.WithInstance(sqlDB, &pgx5.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "pgx5", dbDriver)
	if err != nil {
		return err
	}

	defer migrator.Close()

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Database struct {
	*pgxpool.Pool
}

func NewDatabase(ctx context.Context, DSN string) (*Database, error) {
	pool, err := pgxpool.New(ctx, DSN)
	if err != nil {
		return nil, err
	}
	db := Database{
		pool,
	}
	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}
	err = db.Migrate(ctx)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *Database) Migrate(ctx context.Context) error {
	sqlDB := stdlib.OpenDB(*db.Config().ConnConfig)
	defer sqlDB.Close()

	migrationsDir := "./migrations"
	// if err := goose.ResetContext(ctx, db.DB, migrationsDir); err != nil {
	// 	return err
	// }
	if err := goose.UpContext(ctx, sqlDB, migrationsDir); err != nil {
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}

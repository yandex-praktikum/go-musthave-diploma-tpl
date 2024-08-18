package database

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Database struct {
	*sql.DB
}

func NewDatabase(ctx context.Context, DSN string) (*Database, error) {
	sqlDB, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, err
	}
	db := Database{
		sqlDB,
	}
	err = db.PingContext(ctx)
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
	migrationsDir := "./migrations"
	if err := goose.UpContext(ctx, db.DB, migrationsDir); err != nil {
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}

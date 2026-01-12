package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func Apply(ctx context.Context, db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(embedMigrations)

	const migrationsDir = "sql"

	log.Printf("applying goose migrations from %s", migrationsDir)
	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

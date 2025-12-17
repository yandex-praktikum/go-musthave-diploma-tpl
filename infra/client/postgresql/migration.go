package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
	"log"
	"os"
)

func Migrations(db *sql.DB, migrationDir string) error {

	// Устанавливаем миграции из указанной директории
	goose.SetBaseFS(os.DirFS(migrationDir))

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Применяем миграции до последней версии
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Migrations applied successfully!")
	return nil
}

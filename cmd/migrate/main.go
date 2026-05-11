package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: migrate <up|down> <database-uri>")
		os.Exit(1)
	}

	command := os.Args[1]
	dsn := os.Args[2]

	// Открываем подключение
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Создаем драйвер базы данных
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Printf("Error creating driver: %v\n", err)
		os.Exit(1)
	}

	// Создаем источник миграций
	source, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		fmt.Printf("Error creating source: %v\n", err)
		os.Exit(1)
	}

	// Создаем экземпляр migrate
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		fmt.Printf("Error creating migrate instance: %v\n", err)
		os.Exit(1)
	}

	// Выполняем команду
	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("Error running up migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations up completed successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("Error running down migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations down completed successfully")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

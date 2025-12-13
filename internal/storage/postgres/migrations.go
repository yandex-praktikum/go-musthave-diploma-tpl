package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ApplyMigrations применяет миграции из папки migrations (публичная функция)
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	return applyMigrations(ctx, db)
}

// RollbackMigrations откатывает миграции (публичная функция)
func RollbackMigrations(ctx context.Context, db *sql.DB, targetVersion int64) error {
	return rollbackMigrations(ctx, db, targetVersion)
}

// applyMigrations применяет миграции из папки migrations
func applyMigrations(ctx context.Context, db *sql.DB) error {
	// Читаем файлы миграций из папки migrations
	migrations, err := readMigrations("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Начинаем транзакцию с контекстом
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем таблицу для отслеживания миграций если её нет
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем список примененных миграций
	appliedVersions := make(map[int64]bool)
	rows, err := tx.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedVersions[version] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating applied migrations: %w", err)
	}

	// Применяем миграции по порядку
	for _, migration := range migrations {
		if appliedVersions[migration.Version] {
			continue // Миграция уже применена
		}

		// Выполняем UP-миграцию
		_, err := tx.ExecContext(ctx, migration.UpSQL)
		if err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Записываем версию в таблицу миграций
		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		fmt.Printf("Applied migration: %s\n", migration.Name)
	}

	return tx.Commit()
}

// rollbackMigrations откатывает миграции
func rollbackMigrations(ctx context.Context, db *sql.DB, targetVersion int64) error {
	migrations, err := readMigrations("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Сортируем в обратном порядке для отката
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version > migrations[j].Version
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Получаем текущую версию
	var currentVersion int64
	err = tx.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Откатываем миграции
	for _, migration := range migrations {
		if migration.Version <= targetVersion {
			break // Достигли целевой версии
		}

		if migration.Version > currentVersion {
			continue // Эта миграция еще не применена
		}

		if migration.DownSQL == "" {
			return fmt.Errorf("no down migration for version %d", migration.Version)
		}

		// Выполняем DOWN-миграцию
		_, err := tx.ExecContext(ctx, migration.DownSQL)
		if err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
		}

		// Удаляем запись о миграции
		_, err = tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.Version)
		if err != nil {
			return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
		}

		fmt.Printf("Rolled back migration: %s\n", migration.Name)
	}

	return tx.Commit()
}

// Migration представляет одну миграцию
type Migration struct {
	Version int64
	Name    string
	UpSQL   string
	DownSQL string
}

// readMigrations читает миграции из указанной папки
func readMigrations(migrationsDir string) ([]Migration, error) {
	var migrations []Migration

	// Читаем файлы в папке миграций
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Собираем информацию о миграциях
	migrationFiles := make(map[string]string) // version -> {up|down} -> filepath

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := filepath.Ext(filename)
		if ext != ".sql" {
			continue
		}

		baseName := strings.TrimSuffix(filename, ext)
		parts := strings.Split(baseName, "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue // Пропускаем файлы без числовой версии
		}

		// Определяем тип миграции (up или down)
		var migrationType string
		if strings.HasSuffix(baseName, ".up") {
			migrationType = "up"
		} else if strings.HasSuffix(baseName, ".down") {
			migrationType = "down"
		} else {
			continue // Пропускаем файлы без указания типа
		}

		key := fmt.Sprintf("%d_%s", version, migrationType)
		migrationFiles[key] = filepath.Join(migrationsDir, filename)
	}

	// Собираем миграции
	versions := make(map[int64]bool)
	for key := range migrationFiles {
		parts := strings.Split(key, "_")
		if len(parts) < 2 {
			continue
		}
		version, _ := strconv.ParseInt(parts[0], 10, 64)
		versions[version] = true
	}

	// Сортируем версии
	var sortedVersions []int64
	for version := range versions {
		sortedVersions = append(sortedVersions, version)
	}
	sort.Slice(sortedVersions, func(i, j int) bool {
		return sortedVersions[i] < sortedVersions[j]
	})

	// Читаем SQL для каждой миграции
	for _, version := range sortedVersions {
		upKey := fmt.Sprintf("%d_up", version)
		downKey := fmt.Sprintf("%d_down", version)

		upFile, hasUp := migrationFiles[upKey]
		downFile, hasDown := migrationFiles[downKey]

		if !hasUp {
			return nil, fmt.Errorf("missing up migration for version %d", version)
		}

		upSQL, err := os.ReadFile(upFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read up migration %d: %w", version, err)
		}

		var downSQL string
		if hasDown {
			downData, err := os.ReadFile(downFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read down migration %d: %w", version, err)
			}
			downSQL = string(downData)
		}

		migration := Migration{
			Version: version,
			Name:    filepath.Base(upFile),
			UpSQL:   string(upSQL),
			DownSQL: downSQL,
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

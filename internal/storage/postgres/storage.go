package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/error/pgerrors"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/models"
	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/retry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DatabaseStorage интерфейс для проверки подключения к БД
type DatabaseStorage interface {
	Ping(ctx context.Context) error
	Close() error
}

// Storage интерфейс для работы с метриками
type Storage interface {
	DatabaseStorage
	// SetGauge(ctx context.Context, name string, value float64)
	// IncrementCounter(ctx context.Context, name string, delta int64)
	// GetMetric(ctx context.Context, name string, metricType models.MetricType) (interface{}, bool)
	// GetAllMetrics(ctx context.Context) (map[string]float64, map[string]int64)
	// GetMetricForJSON(ctx context.Context, name string, metricType models.MetricType) models.Metrics
	// SaveToFile(ctx context.Context, filename string) error
	// LoadFromFile(ctx context.Context, filename string) error
	// UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error
}

// PostgresStorage реализация Storage для PostgreSQL
type PostgresStorage struct {
	db         *sql.DB
	classifier retry.ErrorClassifier
}

// NewPostgresStorage создает новое подключение к PostgreSQL
func NewPostgresStorage(ctx context.Context, connectionString string) (Storage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем подключение с использованием переданного контекста
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Применяем миграции с использованием переданного контекста
	if err := ApplyMigrations(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &PostgresStorage{
		db:         db,
		classifier: pgerrors.NewPostgresErrorClassifier(),
	}, nil
}

// SaveToFile - для совместимости с интерфейсом
func (s *PostgresStorage) SaveToFile(ctx context.Context, filename string) error {
	return nil
}

// LoadFromFile - для совместимости с интерфейсом
func (s *PostgresStorage) LoadFromFile(ctx context.Context, filename string) error {
	return nil
}

// Close закрывает соединение с БД
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// Ping проверяет соединение с БД
func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// // UpdateMetricsBatch обновляет метрики батчем в транзакции с повторными попытками
// func (s *PostgresStorage) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
// 	operation := func() error {
// 		return s.updateMetricsBatchTx(ctx, metrics)
// 	}

// 	return retry.WithRetry(ctx, retry.DefaultRetryConfig, operation, s.classifier)
// }

// // updateMetricsBatchTx внутренняя функция для обновления метрик в транзакции
// func (s *PostgresStorage) updateMetricsBatchTx(ctx context.Context, metrics []models.Metrics) error {
// 	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
// 	defer cancel()

// 	// Начинаем транзакцию
// 	tx, err := s.db.BeginTx(queryCtx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()

// 	// Подготавливаем запросы для gauge и counter
// 	gaugeStmt, err := tx.PrepareContext(queryCtx, `
// 		INSERT INTO gauge_metrics (name, value, updated_at)
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (name)
// 		DO UPDATE SET value = $2, updated_at = $3
// 	`)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare gauge statement: %w", err)
// 	}
// 	defer gaugeStmt.Close()

// 	counterStmt, err := tx.PrepareContext(queryCtx, `
// 		INSERT INTO counter_metrics (name, value, updated_at)
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (name)
// 		DO UPDATE SET value = counter_metrics.value + $2, updated_at = $3
// 	`)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare counter statement: %w", err)
// 	}
// 	defer counterStmt.Close()

// 	// Обрабатываем каждую метрику
// 	for _, metric := range metrics {
// 		switch metric.MType {
// 		case "gauge":
// 			if metric.Value == nil {
// 				continue
// 			}
// 			_, err := gaugeStmt.ExecContext(queryCtx, metric.ID, *metric.Value, time.Now())
// 			if err != nil {
// 				return fmt.Errorf("failed to update gauge metric %s: %w", metric.ID, err)
// 			}

// 		case "counter":
// 			if metric.Delta == nil {
// 				continue
// 			}
// 			_, err := counterStmt.ExecContext(queryCtx, metric.ID, *metric.Delta, time.Now())
// 			if err != nil {
// 				return fmt.Errorf("failed to update counter metric %s: %w", metric.ID, err)
// 			}
// 		}
// 	}

// 	// Коммитим транзакцию
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return nil
// }

// UserStorage интерфейс для работы с пользователями
type UserStorage interface {
	CreateUser(ctx context.Context, login, hashedPassword string) error
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	GetUserByID(ctx context.Context, id int) (models.User, error)
	UserExists(ctx context.Context, login string) (bool, error)
}

// CreateUser создает нового пользователя
func (s *PostgresStorage) CreateUser(ctx context.Context, login, hashedPassword string) error {
	query := `
        INSERT INTO users (login, password, created_at, updated_at) 
        VALUES ($1, $2, $3, $4)
    `

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, login, hashedPassword, now, now)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		if strings.Contains(err.Error(), "duplicate key value") ||
			strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("user with login '%s' already exists", login)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	var user models.User

	query := `
        SELECT id, login, password, created_at, updated_at 
        FROM users 
        WHERE login = $1
    `

	err := s.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, sql.ErrNoRows
		}
		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *PostgresStorage) GetUserByID(ctx context.Context, id int) (models.User, error) {
	var user models.User

	query := `
        SELECT id, login, password, created_at, updated_at 
        FROM users 
        WHERE id = $1
    `

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// UserExists проверяет существование пользователя по логину
func (s *PostgresStorage) UserExists(ctx context.Context, login string) (bool, error) {
	var exists bool

	query := `
        SELECT EXISTS(
            SELECT 1 FROM users WHERE login = $1
        )
    `

	err := s.db.QueryRowContext(ctx, query, login).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

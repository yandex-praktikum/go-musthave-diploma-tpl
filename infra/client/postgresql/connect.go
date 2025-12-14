package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"strings"
	"time"

	"os"
)

// ConnectionConfig содержит настройки подключения
type ConnectionConfig struct {
	Host              string
	Port              string
	Timeout           time.Duration
	MaxRetries        int
	RetryInterval     time.Duration
	BackoffMultiplier float64
}

// TCPChecker интерфейс для проверки TCP-соединения
type TCPChecker interface {
	Check(host, port string, timeout time.Duration) error
}

// RealTCPChecker реализация TCPChecker для реальных подключений
type RealTCPChecker struct{}

// Check проверяет доступность TCP-порта
func (r *RealTCPChecker) Check(host, port string, timeout time.Duration) error {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", address, err)
	}
	conn.Close()
	return nil
}

// RetryConfig конфигурация повторных попыток
type RetryConfig struct {
	MaxRetries        int
	RetryInterval     time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig возвращает конфигурацию по умолчанию
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		RetryInterval:     2 * time.Second,
		BackoffMultiplier: 1.5,
	}
}

// extractHostPort извлекает хост и порт из строки подключения
func extractHostPort(config string) (string, string, error) {
	if strings.Contains(config, "://") {
		start := strings.Index(config, "@")
		if start == -1 {
			return "", "", fmt.Errorf("invalid connection URL format")
		}
		end := strings.Index(config[start+1:], "/")
		if end == -1 {
			return "", "", fmt.Errorf("invalid connection URL format")
		}

		hostPort := config[start+1 : start+1+end]
		parts := strings.Split(hostPort, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid host:port format")
		}
		return parts[0], parts[1], nil
	}

	var host, port string
	pairs := strings.Fields(config)
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "host":
			host = kv[1]
		case "port":
			port = kv[1]
		}
	}

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	return host, port, nil
}

// retryOperation выполняет операцию с повторными попытками
func retryOperation(operation func() error, config RetryConfig) error {
	var lastErr error
	currentInterval := config.RetryInterval

	for i := 0; i <= config.MaxRetries; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil // Успех
		}

		// Если это последняя попытка, не ждем
		if i == config.MaxRetries {
			break
		}

		fmt.Printf("Connection attempt %d/%d failed: %v. Retrying in %v...\n",
			i+1, config.MaxRetries, lastErr, currentInterval)

		time.Sleep(currentInterval)

		// Увеличиваем интервал для следующей попытки (exponential backoff)
		currentInterval = time.Duration(float64(currentInterval) * config.BackoffMultiplier)
	}

	return fmt.Errorf("after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// SafeConn создает пул подключений к базе данных с проверками доступности и повторными попытками
func SafeConn(config string) *pgxpool.Pool {
	return ConnWithRetry(config, &RealTCPChecker{}, 5*time.Second, DefaultRetryConfig())
}

// ConnWithRetry создает пул подключений с кастомными настройками и повторными попытками
func ConnWithRetry(config string, checker TCPChecker, timeout time.Duration, retryConfig RetryConfig) *pgxpool.Pool {
	host, port, err := extractHostPort(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse connection string: %v\n", err)
		//os.Exit(1)
	}

	var pool *pgxpool.Pool

	// Попытка подключения с повторными попытками
	err = retryOperation(func() error {
		// Проверка TCP-доступности
		if tcpErr := checker.Check(host, port, timeout); tcpErr != nil {
			return fmt.Errorf("TCP check failed: %w", tcpErr)
		}

		// Подключение к базе данных
		var connErr error
		pool, connErr = pgxpool.New(context.Background(), config)
		if connErr != nil {
			return fmt.Errorf("database connection failed: %w", connErr)
		}

		// Проверка через ping
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if pingErr := pool.Ping(ctx); pingErr != nil {
			pool.Close()
			return fmt.Errorf("database ping failed: %w", pingErr)
		}

		return nil
	}, retryConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to establish database connection: %v\n", err)
		//os.Exit(1)
	}

	return pool
}

// ConnWithConfig создает пул подключений с кастомными настройками (обратная совместимость)
func ConnWithConfig(config string, checker TCPChecker, timeout time.Duration) *pgxpool.Pool {
	return ConnWithRetry(config, checker, timeout, DefaultRetryConfig())
}

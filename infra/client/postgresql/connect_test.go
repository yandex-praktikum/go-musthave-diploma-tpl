package postgresql

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// MockTCPCheckerWithRetries мок для тестирования повторных попыток TCP-проверки
type MockTCPCheckerWithRetries struct {
	FailCount    int
	CurrentCount int
	LastHost     string
	LastPort     string
}

func (m *MockTCPCheckerWithRetries) Check(host, port string, timeout time.Duration) error {
	m.LastHost = host
	m.LastPort = port
	m.CurrentCount++

	if m.CurrentCount <= m.FailCount {
		return fmt.Errorf("mock TCP connection failed (attempt %d)", m.CurrentCount)
	}
	return nil
}

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name:     "URL format",
			config:   "postgres://pgx_md5:secret@localhost:5432/pgx_test?sslmode=disable",
			wantHost: "localhost",
			wantPort: "5432",
			wantErr:  false,
		},
		{
			name:     "Key-value format",
			config:   "host=localhost port=5432 user=pgx_md5 dbname=pgx_test",
			wantHost: "localhost",
			wantPort: "5432",
			wantErr:  false,
		},
		{
			name:     "Key-value with default port",
			config:   "host=127.0.0.1 user=pgx_md5 dbname=pgx_test",
			wantHost: "127.0.0.1",
			wantPort: "5432",
			wantErr:  false,
		},
		{
			name:     "Invalid URL format",
			config:   "postgres://pgx_md5:secret@localhost/pgx_test",
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := extractHostPort(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractHostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if host != tt.wantHost {
				t.Errorf("extractHostPort() host = %v, want %v", host, tt.wantHost)
			}

			if port != tt.wantPort {
				t.Errorf("extractHostPort() port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestRetryOperation_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	maxCalls := 2

	operation := func() error {
		callCount++
		if callCount < maxCalls {
			return errors.New("временный сбой")
		}
		return nil
	}

	retryConfig := RetryConfig{
		MaxRetries:        3,
		RetryInterval:     10 * time.Millisecond,
		BackoffMultiplier: 1.0,
	}

	err := retryOperation(operation, retryConfig)

	if err != nil {
		t.Errorf("Ожидался успех после повторов, получена ошибка: %v", err)
	}

	if callCount != maxCalls {
		t.Errorf("Expected %d calls, got %d", maxCalls, callCount)
	}
}

func TestRetryOperation_MaxRetriesExceeded(t *testing.T) {
	callCount := 0

	operation := func() error {
		callCount++
		return errors.New("постоянный сбой")
	}

	retryConfig := RetryConfig{
		MaxRetries:        2,
		RetryInterval:     10 * time.Millisecond,
		BackoffMultiplier: 1.0,
	}

	err := retryOperation(operation, retryConfig)

	if err == nil {
		t.Error("Ожидалась ошибка после превышения максимального количества повторов, получен nil")
	}

	expectedCalls := retryConfig.MaxRetries + 1
	if callCount != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, callCount)
	}

	if !strings.Contains(err.Error(), "after 3 attempts") {
		t.Errorf("Error message should contain attempt count, got: %v", err)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.RetryInterval != 2*time.Second {
		t.Errorf("Expected RetryInterval=2s, got %v", config.RetryInterval)
	}

	if config.BackoffMultiplier != 1.5 {
		t.Errorf("Expected BackoffMultiplier=1.5, got %v", config.BackoffMultiplier)
	}
}

func TestRetryOperation_Backoff(t *testing.T) {
	var intervals []time.Duration
	startTime := time.Now()

	operation := func() error {
		elapsed := time.Since(startTime)
		intervals = append(intervals, elapsed)
		return errors.New("всегда сбой")
	}

	retryConfig := RetryConfig{
		MaxRetries:        2,
		RetryInterval:     50 * time.Millisecond,
		BackoffMultiplier: 2.0, // Удваиваем интервал каждый раз
	}

	_ = retryOperation(operation, retryConfig)

	// Проверяем, что интервалы увеличиваются
	if len(intervals) >= 3 {
		firstInterval := intervals[1] - intervals[0]
		secondInterval := intervals[2] - intervals[1]

		// Второй интервал должен быть примерно в 2 раза больше первого
		ratio := float64(secondInterval) / float64(firstInterval)
		if ratio < 1.8 || ratio > 2.2 {
			t.Errorf("Expected backoff multiplier ~2.0, got %f", ratio)
		}
	}
}

// BenchmarkRetryOperation бенчмарк для операции с повторными попытками
func BenchmarkRetryOperation(b *testing.B) {
	retryConfig := RetryConfig{
		MaxRetries:        3,
		RetryInterval:     1 * time.Millisecond,
		BackoffMultiplier: 1.0,
	}

	operation := func() error {
		return nil // Всегда успех
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = retryOperation(operation, retryConfig)
	}
}

// BenchmarkExtractHostPort бенчмарк для функции извлечения хоста и порта
func BenchmarkExtractHostPort(b *testing.B) {
	config := "postgres://user:pass@localhost:5432/mydb"

	for i := 0; i < b.N; i++ {
		_, _, _ = extractHostPort(config)
	}
}

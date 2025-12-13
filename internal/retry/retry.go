package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/IvanDolgov/go-musthave-diploma-tpl/internal/logger"
	"go.uber.org/zap"
)

// ErrorClassifier интерфейс для классификации ошибок
type ErrorClassifier interface {
	Classify(err error) ErrorClassification
}

// ErrorClassification тип для классификации ошибок
type ErrorClassification int

const (
	// NonRetriable - операцию не следует повторять
	NonRetriable ErrorClassification = iota

	// Retriable - операцию можно повторить
	Retriable
)

// RetryConfig конфигурация для повторных попыток
type RetryConfig struct {
	MaxAttempts int
	Delays      []time.Duration
}

// DefaultRetryConfig стандартная конфигурация (1s, 3s, 5s)
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
}

// RetriableFunc функция, которую нужно выполнить с повторными попытками
type RetriableFunc func() error

// WithRetry выполняет функцию с повторными попытками для retriable-ошибок
func WithRetry(ctx context.Context, config RetryConfig, fn RetriableFunc, classifier ErrorClassifier) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Выполняем функцию
		err := fn()
		if err == nil {
			if attempt > 0 {
				logger.LogRetrySuccess("operation", attempt+1)
			}
			return nil // Успех
		}

		lastErr = err

		// Проверяем, является ли ошибка retriable
		isRetriable := false
		if classifier != nil {
			classification := classifier.Classify(err)
			isRetriable = (classification == Retriable)
		} else {
			// Если классификатор не передан, считаем все сетевые ошибки retriable
			isRetriable = isNetworkError(err)
		}

		if !isRetriable {
			logger.Log.Error("Non-retriable error encountered", zap.Error(err))
			return fmt.Errorf("non-retriable error: %w", err)
		}

		// Логируем попытку повтора
		delay := getDelay(config, attempt)
		logger.LogRetryAttempt(attempt+1, delay, err)

		// Если это последняя попытка, выходим
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Ждём перед следующей попыткой
		select {
		case <-ctx.Done():
			logger.Log.Error("Retry cancelled by context", zap.Error(ctx.Err()))
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Продолжаем
		}
	}

	logger.LogRetryFailure("operation", config.MaxAttempts, lastErr)
	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// getDelay возвращает задержку для текущей попытки
func getDelay(config RetryConfig, attempt int) time.Duration {
	if attempt < len(config.Delays) {
		return config.Delays[attempt]
	}
	// Если попыток больше, чем задержек, используем последнюю задержку
	return config.Delays[len(config.Delays)-1]
}

// isNetworkError проверяет, является ли ошибка сетевой или временной
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Проверяем стандартные сетевые ошибки через errors.As
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Сетевые ошибки с таймаутом считаем retriable
		if netErr.Timeout() {
			return true
		}
	}

	// Проверяем системные ошибки через errors.As
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		// Конкретные syscall ошибки, которые являются сетевыми
		switch syscallErr {
		case syscall.ECONNREFUSED, // Connection refused
			syscall.ECONNRESET,   // Connection reset by peer
			syscall.ETIMEDOUT,    // Connection timed out
			syscall.EHOSTUNREACH, // No route to host
			syscall.ENETUNREACH,  // Network is unreachable
			syscall.EAGAIN,       // Resource temporarily unavailable
			syscall.ECONNABORTED: // Software caused connection abort
			return true
		}
	}

	// Проверяем DNS ошибки
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// DNS ошибки обычно временные и могут быть retriable
		if dnsErr.IsTemporary || dnsErr.Timeout() {
			return true
		}
	}

	// Проверяем ошибки операций с файлами (могут быть связаны с сетью)
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		// Рекурсивно проверяем обернутую ошибку
		return isNetworkError(pathErr.Err)
	}

	// Проверяем конкретные типы ошибок через errors.Is
	switch {
	case errors.Is(err, net.ErrClosed), // Use of closed network connection
		errors.Is(err, os.ErrDeadlineExceeded),   // I/O timeout
		errors.Is(err, context.DeadlineExceeded), // Context deadline exceeded
		errors.Is(err, context.Canceled):         // Context canceled
		return true
	}

	// Рекурсивно проверяем обернутые ошибки
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		return isNetworkError(unwrapped)
	}

	return false
}

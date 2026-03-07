package backoff

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPgBackoff_WithRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("Success on first attempt", func(t *testing.T) {
		backoff := NewPgBackoff(3, 1*time.Second)
		calls := 0
		err := backoff.WithRetry(ctx, func() error {
			calls++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("Retry on retriable error and succeed", func(t *testing.T) {
		backoff := NewPgBackoff(3, 500*time.Millisecond)
		calls := 0
		// Эмулируем ошибку дедлока (Retriable)
		retriableErr := &pgconn.PgError{Code: pgerrcode.DeadlockDetected}

		err := backoff.WithRetry(ctx, func() error {
			calls++
			if calls < 2 {
				return retriableErr
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("Stop on non-retriable error", func(t *testing.T) {
		backoff := NewPgBackoff(5, 1*time.Second)
		calls := 0
		// Ошибка уникальности (Non-retriable)
		fatalErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

		err := backoff.WithRetry(ctx, func() error {
			calls++
			return fatalErr
		})
		assert.Error(t, err)
		assert.Equal(t, 1, calls, "Should not retry on UniqueViolation")
	})

	t.Run("Context cancellation during wait", func(t *testing.T) {
		backoff := NewPgBackoff(3, 2*time.Second)
		cancelCtx, cancel := context.WithCancel(ctx)

		calls := 0
		go func() {
			time.Sleep(150 * time.Millisecond)
			cancel()
		}()

		err := backoff.WithRetry(cancelCtx, func() error {
			calls++
			return &pgconn.PgError{Code: pgerrcode.SerializationFailure}
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})
}

func TestErrorClassifiers(t *testing.T) {
	t.Run("IsTimeout", func(t *testing.T) {
		assert.False(t, IsTimeout(nil))
		assert.True(t, IsTimeout(context.DeadlineExceeded))

		netErr := &net.DNSError{IsTimeout: true}
		assert.True(t, IsTimeout(netErr))
	})

	t.Run("IsNetworkError", func(t *testing.T) {
		assert.True(t, IsNetworkError(&net.OpError{}))
		assert.True(t, IsNetworkError(&pgconn.ConnectError{}))
		// Класс 08 - Connection Exception
		assert.True(t, IsNetworkError(&pgconn.PgError{Code: "08001"}))
		assert.False(t, IsNetworkError(errors.New("generic")))
	})

	t.Run("IsRetriable_Table", func(t *testing.T) {
		tests := []struct {
			code     string
			expected bool
		}{
			{pgerrcode.DeadlockDetected, true},
			{pgerrcode.UniqueViolation, false},
			{pgerrcode.SyntaxError, false},
			{pgerrcode.AdminShutdown, true},
			{pgerrcode.DataException, false},
		}

		for _, tt := range tests {
			err := &pgconn.PgError{Code: tt.code}
			assert.Equal(t, tt.expected, IsRetriable(err), "Code: %s", tt.code)
		}
	})
}

func TestNewPgBackoff_Limits(t *testing.T) {
	// Проверка коррекции отрицательных попыток
	b1 := NewPgBackoff(0, 10*time.Millisecond)
	assert.Equal(t, 1, b1.attempt)

	// Проверка минимального maxSleep
	b2 := NewPgBackoff(3, 10*time.Millisecond)
	assert.True(t, b2.factor > 1.0)
}

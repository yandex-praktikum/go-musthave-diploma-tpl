package backoff

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=backoff_mock.go -package=backoff

type RetryableOperation interface {
	Execute() error
}

const minDuration time.Duration = 100 * time.Millisecond

type PgBackoff struct {
	attempt      int
	waitDuration time.Duration // Начальная задержка (напр. 100ms)
	factor       float64       // Множитель (обычно 2.0)
}

func NewPgBackoff(attempt int, maxSleep time.Duration) *PgBackoff {
	if attempt <= 0 {
		attempt = 1
	}
	if maxSleep < minDuration {
		maxSleep = 2 * minDuration
	}
	return &PgBackoff{
		attempt:      attempt,
		waitDuration: minDuration,
		factor:       math.Pow(float64(maxSleep)/float64(minDuration), 1.0/float64(attempt)),
	}
}

func (b *PgBackoff) WithRetry(ctx context.Context, cb func() error) error {
	errs := []error{}

	remaining := b.attempt
	currentWait := b.waitDuration

	for i := 0; i < b.attempt; i++ {
		err := cb()
		if err == nil {
			return nil
		}

		errs = append(errs, err)
		remaining--

		if remaining <= 0 || !IsRetriable(err) {
			break
		}

		jitter := time.Duration(float64(currentWait) * 0.1 * (2*rand.Float64() - 1))

		slog.Info("Retry sleep",
			slog.Any("err", err),
			slog.Duration("delay", currentWait+jitter),
			slog.Int("attempt", remaining),
		)

		timer := time.NewTimer(currentWait + jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(append(errs, ctx.Err())...)
		case <-timer.C:
			currentWait = time.Duration(float64(currentWait) * b.factor)
		}

	}
	return errors.Join(errs...)
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}

func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}
	return false
}

func IsRetriable(err error) bool {
	if IsNetworkError(err) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		// ----  Нельзя ретраить  ----
		// Класс 22 - Ошибки данных
		case pgerrcode.DataException,
			pgerrcode.NullValueNotAllowedDataException:
			return false

		// Класс 23 - Нарушение ограничений целостности
		case pgerrcode.IntegrityConstraintViolation,
			pgerrcode.RestrictViolation,
			pgerrcode.NotNullViolation,
			pgerrcode.ForeignKeyViolation,
			pgerrcode.UniqueViolation,
			pgerrcode.CheckViolation:
			return false

		// Класс 42 - Синтаксические ошибки
		case pgerrcode.SyntaxErrorOrAccessRuleViolation,
			pgerrcode.SyntaxError,
			pgerrcode.UndefinedColumn,
			pgerrcode.UndefinedTable,
			pgerrcode.UndefinedFunction:
			return false

		// ----  Можно ретраить  ----
		// Класс 40 - Откат транзакции
		case pgerrcode.TransactionRollback, // 40000
			pgerrcode.SerializationFailure, // 40001
			pgerrcode.DeadlockDetected:     // 40P01
			return true

		// Класс 57 - Ошибка оператора
		case pgerrcode.CannotConnectNow, // 57P03
			pgerrcode.AdminShutdown, // 57P01
			pgerrcode.CrashShutdown: // 57P02
			return true
		}
	}
	return false
}

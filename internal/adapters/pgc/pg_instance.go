package pgc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc/backoff"
	"gmart/internal/adapters/pgc/fcounter"
	"gmart/migrations"

	"golang.org/x/sys/cpu"

	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	goose "github.com/pressly/goose/v3"
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=pg_instance_mock.go  -package=pgc
//go:generate $GOPATH/bin/mockgen                 -destination=pgx_mock.go          -package=pgc github.com/jackc/pgx/v5 Tx,Row,BatchResults

// PgxPoolIface интерфейс, который ограничивает методы, передаваемые в коллбек
type PgxPoolIface interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Stat() *pgxpool.Stat
}

// PgxTxIface интерфейс, который ограничивает методы, передаваемые в коллбек
type PgxTxIface interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	LargeObjects() pgx.LargeObjects
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PgInstance коннектор, предохраняющий БД от дополнительной нагрузки, когда БД "плохо"
// и предохраняющий приложение от каскадного сбоя при проблемах с БД
type PgInstance interface {
	// RunMigrations запуск миграций на инстансе
	RunMigrations(ctx context.Context) error

	// String стрингер "hostname:port/database"
	String() string

	// Close перевод хэндла в IsClosed && !IsReady режим
	Close() error

	// Ping Проверяет работоспособность БД.
	// По результатам работы устанавливается флаг isReady и сбрасывается счетчик h.failures
	Ping(ctx context.Context) error

	// Tx вызвает коллбек и передает ему открытую транзакцию PostgreSQL
	Tx(ctx context.Context, cb func(ctx context.Context, tx PgxTxIface) error) error

	// PgPool вызывает коллбек и передает ему целяй пул коннектов к PostgreSQL
	PgPool(ctx context.Context, cb func(ctx context.Context, pool PgxPoolIface) error) error
}

const ProbeInterval int64 = 5

// Отрываем PgInterface от pgxpool.Pool таким враппером
type pgxPoolDriverIface interface {
	PgxPoolIface
	AsPool() *pgxpool.Pool
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
}
type pgxPoolWrapper struct {
	*pgxpool.Pool
}

func (w *pgxPoolWrapper) AsPool() *pgxpool.Pool {
	return w.Pool
}

type PgInstanceMetrics interface {
	SetStatus(instance string, online bool)
	IncOfflineEvent(instance string)
	IncRetry(instance string, opType metrics.OpType)
	ObserveLatency(instance string, opType metrics.OpType, duration float64)
}

// PgInstance коннектор, предохраняющий БД от дополнительной нагрузки, когда БД "плохо"
// и предохраняющий приложение от каскадного сбоя при проблемах с БД
type pgInstance struct {
	mu              sync.Mutex
	_               cpu.CacheLinePad
	instanceName    string
	pgPool          pgxPoolDriverIface
	isReady         atomic.Bool
	isClosed        atomic.Bool
	lastCheckResult atomic.Value

	failures  *fcounter.FailureCounter
	repeater  *backoff.PgBackoff
	lastRetry atomic.Int64
	metrics   PgInstanceMetrics
}

func NewPgInstance(ctx context.Context, connString string, metrics PgInstanceMetrics) (*pgInstance, error) {
	pool, err := pgxpool.New(ctx, connString)

	if err != nil {
		return nil, fmt.Errorf("bad connString for pool: %w", err)
	}

	connCfg := pool.Config().ConnConfig

	h := &pgInstance{
		pgPool:       &pgxPoolWrapper{pool},
		instanceName: fmt.Sprintf("%s:%d/%s", connCfg.Host, connCfg.Port, connCfg.Database),

		failures: fcounter.NewFailureCounter(3, 2*time.Second), // подряд 3 ошибки за 2 секунды и надо перевести PgInstance в Offline
		repeater: backoff.NewPgBackoff(3, 3*time.Second),       // 3 запроса к БД в течение 3 сек
		metrics:  metrics,
	}

	h.isClosed.Store(false)
	h.isReady.Store(false)

	h.Ping(ctx)

	return h, nil
}

func (h *pgInstance) RunMigrations(ctx context.Context) error {
	goose.SetBaseFS(migrations.MigrationsDir)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	migrationsPath := "."

	slog.Info("Running migrations",
		slog.String("instance", h.String()),
		slog.String("migrations path", migrationsPath),
	)

	pool := h.pgPool.AsPool()
	if pool == nil {
		return errors.New("it is a mock")
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.UpContext(ctx, db, migrationsPath); err != nil {
		return fmt.Errorf("failed to migrate PgInstance %s: %w", h.String(), err)
	}

	slog.Info("Successfully migrated database",
		slog.String("instance", h.String()),
	)

	return nil
}

// DbName hostname:port/database соединения к БД
func (h *pgInstance) String() string {
	return h.instanceName
}

// IsReady DB готова к работе, если нет проблем с ошибками и коннект не закрыт
func (h *pgInstance) IsReady() bool {
	return !h.isClosed.Load() && h.isReady.Load()
}

// Online перевод хэндла в IsReady режим
func (h *pgInstance) Online() {
	if h.isReady.CompareAndSwap(false, true) {
		if h.metrics != nil {
			h.metrics.SetStatus(h.String(), true)
		}
		slog.Info("The PgInstance has gone Online",
			slog.String("instance", h.String()),
		)
	}
}

// Offline перевод хэндла в !IsReady режим
func (h *pgInstance) Offline() {
	if h.isReady.CompareAndSwap(true, false) {
		h.lastRetry.Store(time.Now().Unix())
		if h.metrics != nil {
			h.metrics.SetStatus(h.String(), false)
			h.metrics.IncOfflineEvent(h.String())
		}
		slog.Info("The PgInstance has gone Offline",
			slog.String("instance", h.String()),
		)
	}
}

// Close перевод хэндла в IsClosed && !IsReady режим
func (h *pgInstance) Close() error {
	if h.isClosed.CompareAndSwap(false, true) {
		h.pgPool.Close()
		slog.Info("The PgInstance closed",
			slog.String("instance", h.String()),
		)
		h.Offline()
	}
	return nil
}

// HandleError реагирует на ошибки
func (h *pgInstance) HandleError(err error) error {
	if err == nil {
		h.failures.Reset()
		h.Online()
		return nil
	}
	if h.failures.Inc(err) {
		h.Offline()
	}

	return err
}

// CanTry определяет одну горутину, которая может попробовать послать запрос в Offline pgxpool.Pool
func (h *pgInstance) CanTry() bool {
	if h.IsReady() {
		return true
	}

	now := time.Now().Unix()
	last := h.lastRetry.Load()

	if now-last >= ProbeInterval {
		return h.lastRetry.CompareAndSwap(last, now)
	}
	return false
}

// Ping Проверяет работоспособность БД.
// По результатам работы устанавливается флаг isReady и сбрасывается счетчик h.failures
func (h *pgInstance) Ping(ctx context.Context) error {
	pCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 1*time.Second)
	defer cancel()

	return h.HandleError(h.pgPool.Ping(pCtx))
}

// Tx вызвает коллбек и передает ему открытую транзакцию PostgreSQL
func (h *pgInstance) Tx(ctx context.Context, cb func(ctx context.Context, tx PgxTxIface) error) error {
	start := time.Now()
	attempt := 0

	err := h.repeater.WithRetry(ctx, func() (err error) {
		if h.metrics != nil && attempt > 0 {
			h.metrics.IncRetry(h.String(), metrics.OpTx)
		}
		attempt++

		var tx pgx.Tx

		defer func() {
			if r := recover(); r != nil {
				slog.Error(
					"panic recovered",
					slog.Any("r", r),
					slog.Any("err", err),
					slog.String("stack", string(debug.Stack())),
				)
				if err != nil {
					err = fmt.Errorf("panic recovered: %w", err)
				} else {
					err = errors.New("panic recovered")
				}
			}

			if err != nil && tx != nil {
				_ = tx.Rollback(context.WithoutCancel(ctx))
			}
		}()

		if !h.CanTry() {
			return errors.New("PgInstance is not ready")
		}

		tx, err = h.pgPool.Begin(ctx)
		if err != nil {
			return h.HandleError(err)
		}

		err = cb(ctx, tx)
		if err != nil {
			return h.HandleError(err)
		}

		err = tx.Commit(context.WithoutCancel(ctx))
		if errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}

		return h.HandleError(err)
	})

	if h.metrics != nil {
		h.metrics.ObserveLatency(h.String(), metrics.OpTx, time.Since(start).Seconds())
	}
	return err
}

// PgPool вызывает коллбек и передает ему целяй пул коннектов к PostgreSQL
func (h *pgInstance) PgPool(ctx context.Context, cb func(ctx context.Context, pool PgxPoolIface) error) error {
	start := time.Now()
	attempt := 0

	err := h.repeater.WithRetry(ctx, func() (err error) {
		if h.metrics != nil && attempt > 0 {
			h.metrics.IncRetry(h.String(), metrics.OpPool)
		}
		attempt++
		defer func() {
			if r := recover(); r != nil {
				slog.Error(
					"panic recovered",
					slog.Any("r", r),
					slog.Any("err", err),
					slog.String("stack", string(debug.Stack())),
				)
				if err != nil {
					err = fmt.Errorf("panic recovered: %w", err)
				} else {
					err = errors.New("panic recovered")
				}
			}
		}()

		if !h.CanTry() {
			return errors.New("PgInstance is not ready")
		}

		err = cb(ctx, h.pgPool)
		return h.HandleError(err)
	})
	if h.metrics != nil {
		h.metrics.ObserveLatency(h.String(), metrics.OpPool, time.Since(start).Seconds())
	}
	return err
}

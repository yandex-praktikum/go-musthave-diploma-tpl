package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc"
	"gmart/internal/domain"
	"gmart/internal/dto"

	"github.com/jackc/pgx/v5"
)

const sqlInsertIntoOrders = `
	WITH ins AS (
		INSERT INTO orders (order_number, user_id, status, uploaded_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (order_number) DO NOTHING
		RETURNING user_id
	)
	SELECT user_id, false AS conflict FROM ins
	UNION ALL
	SELECT user_id, true AS conflict FROM orders WHERE order_number = $1
	LIMIT 1;
`

const sqlSelectOrdersByUserID = `
	SELECT order_number, status, accrual, uploaded_at
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded_at DESC
`

const sqlAcquireNextOrder = `
	UPDATE orders
	SET accrualed_at = $1
	WHERE id = (
		SELECT id
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		  AND (accrualed_at IS NULL OR accrualed_at < $1 - INTERVAL '30 seconds')
		ORDER BY uploaded_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	)
	RETURNING order_number, status;
`
const sqlUpdateOrderStatus = `
	UPDATE orders
	SET status = $2,
	    accrual = $3,
	    accrualed_at = $4 -- r.now()
	WHERE order_number = $1
`

var (
	ErrOrderConflict        = errors.New("order number already uploaded by another user")
	ErrOrderAlreadyUploaded = errors.New("order number already uploaded")
	ErrQueueIsEmpty         = errors.New("queue is empty")
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=repository_mock.go  -package=orders

type OrdersMetrics interface {
	ObserveDB(op metrics.OpType, duration time.Duration)
	IncOrderUpload(status string)                // status: new, conflict, already_uploaded
	IncAcquireAttempt(found bool)                // фиксируем, были ли задачи в очереди
	ObserveListSize(size int)                    // метрика для мониторинга объема данных
	IncOrderFinalized(status domain.OrderStatus) // метрика финализации
}

type OrdersRepo struct {
	pg      pgc.PgInstance
	metrics OrdersMetrics
	now     func() time.Time
}

func NewOrdersRepo(pg pgc.PgInstance, m OrdersMetrics) *OrdersRepo {
	return &OrdersRepo{
		pg:      pg,
		metrics: m,
		now:     time.Now,
	}
}

// Upload загружает новый номер заказа в БД
func (r *OrdersRepo) Upload(ctx context.Context, userID domain.UserID, orderNumber domain.OrderNumber) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpPool, time.Since(start))
		}
	}()

	var ownerID domain.UserID
	var conflict bool

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		// Используем "NEW" как статус по умолчанию
		return pool.QueryRow(ctx, sqlInsertIntoOrders, orderNumber, userID, "NEW", r.now()).Scan(&ownerID, &conflict)
	})

	if err != nil {
		slog.Error("db query failed", "op", "OrdersRepo.Upload", "err", err, "order", orderNumber)
		return fmt.Errorf("upload order fail: %w", err)
	}

	if conflict {
		if ownerID != userID {
			if r.metrics != nil {
				r.metrics.IncOrderUpload("conflict")
			}
			slog.Warn("order conflict", "requested_by", userID, "actual_owner", ownerID, "order", orderNumber)
			return ErrOrderConflict
		}
		if r.metrics != nil {
			r.metrics.IncOrderUpload("already_uploaded")
		}
		return ErrOrderAlreadyUploaded
	}

	if r.metrics != nil {
		r.metrics.IncOrderUpload("new")
	}
	return nil
}

// List выводит список номеров заказов пользователя
func (r *OrdersRepo) List(ctx context.Context, userID domain.UserID) ([]dto.OrderItem, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpPool, time.Since(start))
		}
	}()

	result := make([]dto.OrderItem, 0, 10)

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		rows, err := pool.Query(ctx, sqlSelectOrdersByUserID, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item dto.OrderItem
			if err := rows.Scan(&item.Number, &item.Status, &item.Accrual, &item.UploadedAt); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})

	if err != nil {
		slog.Error("db query failed", "op", "OrdersRepo.List", "err", err, "user_id", userID)
		return nil, fmt.Errorf("list orders fail: %w", err)
	}

	if r.metrics != nil {
		r.metrics.ObserveListSize(len(result))
	}

	return result, nil
}

// AcquireNextOrder резервирует следующий заказ для обработки воркером
func (r *OrdersRepo) AcquireNextOrder(ctx context.Context) (domain.OrderNumber, domain.OrderStatus, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpPool, time.Since(start))
		}
	}()

	var number domain.OrderNumber
	var status domain.OrderStatus

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		return pool.QueryRow(ctx, sqlAcquireNextOrder, r.now()).Scan(&number, &status)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if r.metrics != nil {
				r.metrics.IncAcquireAttempt(false) // Очередь пуста
			}
			return "", "", ErrQueueIsEmpty
		}
		return "", "", fmt.Errorf("acquire order fail: %w", err)
	}

	if r.metrics != nil {
		r.metrics.IncAcquireAttempt(true) // Заказ успешно захвачен
	}
	return number, status, nil
}

// UpdateOrderStatus обновляет статус заказа и сумму начисления.
func (r *OrdersRepo) UpdateOrderStatus(
	ctx context.Context,
	orderNumber domain.OrderNumber,
	orderStatus domain.OrderStatus,
	accrual domain.Amount,
) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.ObserveDB(metrics.OpPool, time.Since(start))
		}
	}()

	err := r.pg.PgPool(ctx, func(ctx context.Context, pool pgc.PgxPoolIface) error {
		tag, err := pool.Exec(ctx, sqlUpdateOrderStatus, orderNumber, orderStatus, accrual, r.now())
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("order %s not found", orderNumber)
		}
		return nil
	})

	if err != nil {
		slog.Error("db query failed",
			slog.String("op", "OrdersRepo.UpdateOrderStatus"),
			slog.Any("err", err),
			slog.Any("orderNumber", orderNumber),
			slog.Any("orderStatus", orderStatus),
		)
		return fmt.Errorf("update order status fail: %w", err)
	}

	// Фиксируем переход в финальный статус
	if orderStatus == "PROCESSED" || orderStatus == "INVALID" {
		if r.metrics != nil {
			r.metrics.IncOrderFinalized(orderStatus)
		}
	}

	return nil
}

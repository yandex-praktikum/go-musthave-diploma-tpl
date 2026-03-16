package loyalty

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

// AddOrder создает заявку на начисление баллов в заказе
func (i *Implementation) AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error {
	tx, err := i.c.Begin(ctx)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	sqlOrder := orderModel{}
	err = tx.QueryRow(ctx, `select id, user_id, order_id from orders where order_id = $1`, order.ID.ID).Scan(&sqlOrder.ID, &sqlOrder.UserID, &sqlOrder.OrderID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return serviceerrors.NewAppError(err)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		switch {
		case uint64(sqlOrder.UserID.Int64) != userID.ID:
			return serviceerrors.NewConflict().Wrap(err, "order belongs other user")
		case uint64(sqlOrder.UserID.Int64) == userID.ID:
			return serviceerrors.NewBadRequest().Wrap(domain.ErrActionCompletedEarly, "")
		}
	}

	_, err = tx.Exec(ctx, "INSERT INTO orders (user_id, order_id, state) VALUES ($1, $2, $3)",
		userID.ID, order.ID.ID, domain.New)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return serviceerrors.NewAppError(err)
	}

	return nil

}

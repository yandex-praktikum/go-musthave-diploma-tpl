package loyalty

import (
	"context"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

func (i *Implementation) AccrualPoints(ctx context.Context, o domain.Order) error {
	tx, err := i.c.Begin(ctx)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				logger.Errorf(ctx, "Rollback transaction unsuccessful: %v", err)
			}
		}
	}()

	_, err = tx.Exec(ctx, `select * from orders where order_id = $1 for update`, o.ID.ID)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	_, err = tx.Exec(ctx, `update orders set amount = $1, currency = $2, state = $3 where order_id = $4`, o.AccrualAmount.Amount,
		o.AccrualAmount.Currency, o.State, o.ID.ID)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	_, err = tx.Exec(ctx, `insert into operations (user_id, order_id, amount, currency, operation)
			values ($1, $2, $3, $4, $5)`, o.UserID.ID, o.ID.ID, o.AccrualAmount.Amount,
		o.AccrualAmount.Currency, domain.Accrual)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return serviceerrors.NewAppError(err)
	}

	return nil
}

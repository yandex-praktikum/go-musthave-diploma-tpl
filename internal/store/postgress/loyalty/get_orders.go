package loyalty

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

func (i *Implementation) GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error) {
	iter, err := i.c.Query(ctx, `select user_id, order_id, state, created_at, currency, amount
			from orders where user_id = $1`, userID.ID)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

	defer iter.Close()

	res := make(domain.Orders, 0)
	for iter.Next() {
		order := orderModel{}
		err = iter.Scan(&order.UserID, &order.OrderID, &order.Status,
			&order.CreatedAt, &order.Currency, &order.AccrualAmount)
		if err != nil {
			return nil, serviceerrors.NewAppError(err)
		}

		state := domain.StateFromString(order.Status.String)
		if state == domain.Invalid {
			logger.Errorf(ctx, " data consistency is broken invalid state: %s, orderID: %d",
				order.Status.String, order.ID.Int64)
		}

		res = append(res, domain.Order{
			ID: domain.ID{
				ID: uint64(order.OrderID.Int64),
			},
			State:     state,
			CreatedAt: order.CreatedAt.Time,
			AccrualAmount: domain.Money{
				Currency: order.Currency.String,
				Amount:   decimal.NewFromFloat(order.AccrualAmount.Float64),
			},
		})
	}

	return res, nil
}

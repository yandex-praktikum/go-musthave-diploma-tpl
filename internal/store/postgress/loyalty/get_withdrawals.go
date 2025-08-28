package loyalty

import (
	"context"
	"database/sql"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type operation struct {
	OrderID   sql.NullInt64   `db:"order_id"`
	Amount    sql.NullFloat64 `db:"amount"`
	CreatedAt sql.NullTime    `db:"created_at"`
}

func (i *Implementation) GetWithdrawals(ctx context.Context, userID domain.ID) ([]domain.Operation, error) {
	iter, err := i.c.Query(ctx, `select order_id, amount, created_at from operations where user_id = $1
  			and operation = 'WITHDRAW'`, userID.ID)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

	res := make([]domain.Operation, 0, 10)
	for iter.Next() {
		o := operation{}
		err = iter.Scan(&o.OrderID, &o.Amount, &o.CreatedAt)
		if err != nil {
			return nil, serviceerrors.NewAppError(err)
		}

		res = append(res, domain.Operation{
			OrderID: domain.ID{
				ID: uint64(o.OrderID.Int64),
			},
			Type: domain.Withdraw,
			Amount: domain.Money{
				Currency: string(domain.GopherMarketBonuses),
				Amount:   decimal.NewFromFloat(o.Amount.Float64),
			},
			CratedAt: o.CreatedAt.Time,
		})
	}

	return res, nil
}

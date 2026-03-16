package loyalty

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

func (i *Implementation) WithdrawPoints(ctx context.Context,
	userID domain.ID, o domain.Operation) error {
	tx, err := i.c.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
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

	_, err = tx.Exec(ctx, `insert into operations (user_id, order_id, amount, currency, operation)
			values ($1, $2, $3, $4, $5)`, userID.ID, o.OrderID.ID, o.Amount.Amount,
		o.Amount.Currency, domain.Withdraw)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	iter, err := tx.Query(ctx, `select operation, sum(amount) as sum from operations where user_id = $1 group by operation`, userID.ID)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	res := domain.Balance{}
	for iter.Next() {
		b := balance{}
		if err = iter.Scan(&b.Operation, &b.Amount); err != nil {
			return serviceerrors.NewAppError(err)
		}

		m := domain.Money{
			Currency: "",
			Amount:   decimal.NewFromFloat(b.Amount.Float64),
		}
		switch b.Operation.String {
		case string(domain.Withdraw):
			res.Withdrawn = m
		case string(domain.Accrual):
			res.Accrual = m
		}
	}
	var rowNum int
	err = tx.QueryRow(ctx, `select  count(*) as sum from operations where order_id = $1   and operation =  'WITHDRAW'`,
		o.OrderID.ID).Scan(&rowNum)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	if rowNum > 1 {
		return serviceerrors.NewBadRequest().Wrap(nil, "try with more than 1 transaction")
	}

	if res.GetCurrent().Amount.LessThan(decimal.Zero) {
		return serviceerrors.NewPaymentRequired().Wrap(nil, "balance less than zero")
	}

	if err = tx.Commit(ctx); err != nil {
		return serviceerrors.NewAppError(err)
	}

	return nil
}

package loyalty

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type Implementation struct {
	c *pgxpool.Pool
}

var accrualTable = `create table if not exists orders (
                                      id           bigint GENERATED ALWAYS AS IDENTITY,
                                      user_id      bigint NOT NULL,
                                      order_id     bigint NOT NULL,                                    
                                      created_at   timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
                                      currency     text NOT NULL default (''),
                                      amount       decimal NOT NULL DEFAULT(0),
                                      state        text,                  
                                      primary key  (user_id, order_id))
`
var operationTable = `create table if not exists operations (
                                         id            bigint GENERATED ALWAYS AS IDENTITY primary key,
                                         order_id      bigint not null,
                                         user_id       bigint not null,
                                         amount        decimal,
                                         currency      text,
                                         operation     text,
                                         created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'));
`

var tables = []string{
	accrualTable,
	operationTable,
}

// NewImplementation ...
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

type orderModel struct {
	ID        sql.NullInt64  `db:"id"`
	UserID    sql.NullInt64  `db:"user_id"`
	OrderID   sql.NullInt64  `db:"order_id"`
	Status    sql.NullString `db:"status"`
	CreatedAt sql.NullTime   `db:"created_at"`
	// TODO поправить
	AccrualAmount sql.NullFloat64 `db:"amount"`
	Currency      sql.NullString  `db:"currency"`
}

func (i *Implementation) UpdateOrderStatus(ctx context.Context, order domain.Order) error {
	_, err := i.c.Exec(ctx, "UPDATE orders SET state = $1 WHERE order_id = $2", order.State, order.ID.ID)
	if err != nil {
		return serviceerrors.NewAppError(err)
	}

	return nil
}

func (i *Implementation) GetOrderForAccruals(ctx context.Context) (domain.Orders, error) {
	iter, err := i.c.Query(ctx, `select user_id, order_id, state, created_at, currency, amount from orders
                            where state = $1 or state = $2 order by orders.created_at desc limit 1`, domain.Processing, domain.New)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

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
			UserID: domain.ID{
				ID: uint64(order.UserID.Int64),
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

type balance struct {
	Currency sql.NullString `db:"currency"`
	// TODO поправить
	Amount    sql.NullFloat64 `db:"amount"`
	Operation sql.NullString  `db:"operation"`
}

func (i *Implementation) GetBalance(ctx context.Context, userID domain.ID) (domain.Balance, error) {
	iter, err := i.c.Query(ctx, `select operation, sum(amount) as sum from operations where user_id = $1 group by operation`, userID.ID)
	if err != nil {
		return domain.Balance{}, serviceerrors.NewAppError(err)
	}

	res := domain.Balance{}
	for iter.Next() {
		b := balance{}
		if err = iter.Scan(&b.Operation, &b.Amount); err != nil {
			return domain.Balance{}, serviceerrors.NewAppError(err)
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

	return res, nil
}

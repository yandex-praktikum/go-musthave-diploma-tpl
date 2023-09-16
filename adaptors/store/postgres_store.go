package store

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	appDb "github.com/k-morozov/go-musthave-diploma-tpl/db"
)

type PostgresStore struct {
	db *sql.DB
}

var _ Store = &PostgresStore{}

func NewPostgresStore(databaseDsn string) (Store, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}

	if err = appDb.InitDB(context.TODO(), db); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (p *PostgresStore) RegisterUser(ctx context.Context, data models.RegisterData) error {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started add new user in store")

	var err error
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if _, err = tx.ExecContext(ctx, appDb.QueryInsertNewUserInUsers, data.ID, data.Username, data.Password); err != nil {
		lg.Err(err).
			Str("user_id", data.ID).
			Str("Username", data.Username).
			Str("error", err.Error()).
			Msg("failed insert new user in db storage")

		return handleUniqueViolation(err)
	}

	if _, err = tx.ExecContext(ctx, appDb.QueryInsertNewUserInUserBalance, data.ID); err != nil {
		lg.Err(err).
			Str("user_id", data.ID).
			Str("error", err.Error()).
			Msg("failed insert new user in db storage")

		return err
	}

	lg.Info().
		Str("user_id", data.ID).
		Str("Username", data.Username).
		Msg("user added to db")

	return nil
}

func (p *PostgresStore) LoginUser(ctx context.Context, data models.LoginData) (string, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started add new user in store")

	rows, err := p.db.QueryContext(ctx, appDb.QueryGetUserIDForExistsUser, data.Username, data.Password)
	if err != nil {
		lg.Err(err).
			Str("Username", data.Username).
			Str("error", err.Error()).
			Msg("failed insert new user in db storage")

		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			return "", err
		}

		lg.Info().
			Str("Username", data.Username).
			Str("user_id", userID).
			Msg("user_id for exists user")

		return userID, nil
	}

	return "", nil
}

func (p *PostgresStore) AddOrder(ctx context.Context, data models.AddOrderData) error {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started add order")

	if _, err := p.db.ExecContext(ctx, appDb.QueryInsertNewOrder, data.UserID, data.OrderID, data.Status); err != nil {
		lg.Warn().Err(err).Msg("Failed add new order")
		return handleUniqueViolation(err)
	}

	lg.Info().
		Str("user_id", data.UserID).
		Str("order_id", data.OrderID).
		Str("status", data.Status.String()).
		Msg("new order added to db")

	return nil
}

func (p *PostgresStore) GetOwnerForOrder(ctx context.Context, data models.GetOwnerForOrderData) (string, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started get owner for order")

	rows, err := p.db.QueryContext(ctx, appDb.QueryGetUserIDForOrder, data.OrderID)
	if err != nil {
		lg.Warn().Err(err).Msg("Failed get owner")
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			return "", err
		}

		lg.Info().
			Str("user_id", userID).
			Str("order_id", data.OrderID).
			Msg("user has order")

		return userID, nil
	}
	return "", nil
}

func (p *PostgresStore) GetOrders(ctx context.Context, data models.GetOrdersData) (models.GetOrdersDataResult, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started get orders for user")
	var result models.GetOrdersDataResult

	rows, err := p.db.QueryContext(ctx, appDb.QueryGetOrdersForUser, data.UserID)
	if err != nil {
		lg.Warn().Err(err).Msg("Failed get owner")
		return models.GetOrdersDataResult{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.OrderData
		if err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return models.GetOrdersDataResult{}, err
		}

		lg.Info().
			Str("order_id", order.Number).
			Msg("found order")

		result.Orders = append(result.Orders, order)
	}

	return result, nil
}

func (p *PostgresStore) GetUserBalance(ctx context.Context, data models.GetUserBalanceData) (models.GetUserBalanceDataResult, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started get user balance")

	var err error
	// @TODO serialisation?
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return models.GetUserBalanceDataResult{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	current, err := getCurrentBalance(ctx, tx, data.UserID, lg)
	if err != nil {
		lg.Err(err).Msg("failed get current balance")
		return models.GetUserBalanceDataResult{}, err
	}
	lg.Info().
		Float64("balance", current).
		Str("user_id", data.UserID).
		Msg("current user balance")

	withdraw, err := getUserSumWithdraw(ctx, tx, data.UserID)
	if err != nil {
		lg.Err(err).Msg("failed get user withdraw")
		return models.GetUserBalanceDataResult{}, err
	}
	lg.Info().
		Float64("withdraw", withdraw).
		Str("user_id", data.UserID).
		Msg("current user withdraw")

	return models.GetUserBalanceDataResult{
		Current:   current,
		Withdrawn: withdraw,
	}, nil
}

func (p *PostgresStore) Withdraw(ctx context.Context, data models.WithdrawData) error {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().
		Str("order_id", data.OrderID).
		Msg("Started add withdraw")

	var err error
	// @TODO serialisation?
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	has, err := userHasMoney(ctx, tx, data.UserID, data.Sum)
	if err != nil {
		return err
	}

	if !has {
		return UserNoMoney{}
	}

	lg.Debug().Msg("user has money")

	_, err = p.db.QueryContext(ctx, appDb.QueryAddWithdraw, data.OrderID, data.UserID, data.Sum)
	if err != nil {
		lg.Warn().Err(err).
			Str("order_id", data.OrderID).
			Str("user_id", data.UserID).
			Msg("Failed add withdraw")
		return err
	}

	lg.Debug().Msg("added withdraw")

	_, err = p.db.QueryContext(ctx, appDb.QueryUpdateUserBalance, data.Sum, data.UserID)
	if err != nil {
		lg.Warn().Err(err).Msg("Failed update balance")
		return err
	}

	lg.Debug().Msg("balance updated")

	return err
}

func (p *PostgresStore) Withdrawals(ctx context.Context, data models.WithdrawalsData) (models.WithdrawsDataResult, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started get withdrawals for user")
	var result models.WithdrawsDataResult

	rows, err := p.db.QueryContext(ctx, appDb.QueryGetWithdrawsForUser, data.UserID)
	if err != nil {
		lg.Warn().Err(err).Msg("Failed get owner")
		return models.WithdrawsDataResult{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdraw models.UserWithdraw
		if err = rows.Scan(&withdraw.OrderID, &withdraw.Sun, &withdraw.ProcessedAt); err != nil {
			return models.WithdrawsDataResult{}, err
		}

		lg.Info().
			Str("order_id", withdraw.OrderID).
			Float64("sun", withdraw.Sun).
			Msg("found withdraw")

		result.Data = append(result.Data, withdraw)
	}

	return result, nil
}

func (p *PostgresStore) GetNewOrders(ctx context.Context, owner, count int) (models.LockNewOrders, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().
		Int("count", count).
		Msg("Started get new orders")

	var err error
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return models.LockNewOrders{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	orders, err := getNewOrders(ctx, tx, lg, count)
	if err != nil {
		lg.Err(err).Msg("failed get new orders")
		return models.LockNewOrders{}, err
	}

	if len(orders) == 0 {
		lg.Info().
			Msg("there is no new orders")
		return models.LockNewOrders{}, nil
	}

	lg.Info().
		Any("orders", orders).
		Msg("new orders without lock")

	if err = lockNewOrders(ctx, tx, owner, orders); err != nil {
		return models.LockNewOrders{}, err
	}

	lg.Info().Msg("successfully locked new order")

	return models.LockNewOrders{Orders: orders}, nil
}

func (p *PostgresStore) RestoreNewOrders(ctx context.Context, owner int) (models.LockNewOrders, error) {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started restore locked orders")

	rows, err := p.db.QueryContext(ctx, appDb.QueryGetLockedNewOrders, owner)
	if err != nil {
		lg.Warn().Err(err).Msg("Failed get orders")
		return models.LockNewOrders{}, err
	}
	defer rows.Close()

	var orders []string
	for rows.Next() {
		var order string
		if err = rows.Scan(&order); err != nil {
			return models.LockNewOrders{}, err
		}

		lg.Info().
			Str("order", order).
			Msg("found order")

		orders = append(orders, order)
	}

	return models.LockNewOrders{Orders: orders}, nil
}

func (p *PostgresStore) ProcessedOrder(ctx context.Context, o models.OrderData) error {
	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Started update processed order")

	if _, err := p.db.ExecContext(ctx, appDb.QueryUpdateProcessedOrder, o.Status, o.Accrual, o.Number); err != nil {
		lg.Warn().Err(err).Msg("Failed update processed order")
		return err
	}

	lg.Info().
		Int("accrual", o.Accrual).
		Str("order_id", o.Number).
		Str("status", o.Status.String()).
		Msg("processed order added to db")

	return nil
}

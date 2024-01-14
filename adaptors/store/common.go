package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog"

	appDb "github.com/k-morozov/go-musthave-diploma-tpl/db"
)

func orderExists(ctx context.Context, tx *sql.Tx, orderID string) (bool, error) {
	rows, err := tx.QueryContext(ctx, appDb.QueryCheckExistsOrderID, orderID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var result bool
		if err = rows.Scan(&result); err != nil {
			return false, err
		}

		return result, nil
	}
	return false, nil
}

func userHasMoney(ctx context.Context, tx *sql.Tx, userID string, sum float64) (bool, error) {
	rows, err := tx.QueryContext(ctx, appDb.QueryCheckUserHasMoney, userID, sum)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var result bool
		if err = rows.Scan(&result); err != nil {
			return false, err
		}

		return result, nil
	}
	return false, nil
}

func getCurrentBalance(ctx context.Context, tx *sql.Tx, userID string, lg zerolog.Logger) (float64, error) {
	lg.Debug().Str("user_id", userID).Msg("start get current balance")
	rows, err := tx.QueryContext(ctx, appDb.QueryGetUserBalance, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var balance float64
		if err = rows.Scan(&balance); err != nil {
			lg.Err(err).Msg("failed scan")
			return 0, err
		}

		lg.Debug().Float64("result", balance).Msg("convert")
		return balance, nil
	}
	return 0, fmt.Errorf("no user for check balance")
}

func getUserSumWithdraw(ctx context.Context, tx *sql.Tx, userID string) (float64, error) {
	rows, err := tx.QueryContext(ctx, appDb.QueryGetUserWithdraw, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var balance float64
		if err = rows.Scan(&balance); err != nil {
			return 0, err
		}

		return balance, nil
	}
	return 0, nil
}

func getNewOrders(ctx context.Context, tx *sql.Tx, lg zerolog.Logger, count int) ([]string, error) {
	lg.Debug().Msg("start get new orders")
	rows, err := tx.QueryContext(ctx, appDb.QueryGetLimitedNewOrders, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orders []string
		var order string
		if err = rows.Scan(&order); err != nil {
			lg.Err(err).Msg("failed scan")
			return nil, err
		}
		orders = append(orders, order)

		return orders, nil
	}
	return []string{}, nil
}

func lockNewOrders(ctx context.Context, tx *sql.Tx, owner int, orders []string) error {
	for _, order := range orders {
		if _, err := tx.ExecContext(ctx, appDb.QueryLockOrders, owner, order); err != nil {
			return err
		}
	}
	return nil
}

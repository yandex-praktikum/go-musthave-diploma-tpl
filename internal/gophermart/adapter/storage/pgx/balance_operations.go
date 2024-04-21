package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgx/v5"
)

func (st *storage) Balance(ctx context.Context, userID int) (*domain.UserBalance, error) {
	st.logger.Infow("storage.Balance", "status", "start")

	var userBalance domain.UserBalance

	if err := st.pPool.QueryRow(ctx,
		` with ins as (
			insert into balance(userId) values ($1) on conflict (userId) do nothing 
			-- явно укащываб список полей, для исключения возможных ошибок
			 returning balanceId, userId, current, withdrawn, release 
			)
			select balanceId, userId, current, withdrawn, release from ins
			union
			select balanceId, userId, current, withdrawn, release from balance
			where userId=$1;`,
		userID).Scan(&userBalance.BalanceId, &userBalance.UserID, &userBalance.Current, &userBalance.Release, &userBalance.Release); err == nil {
		st.logger.Infow("storage.Balance", "status", "success")
		return &userBalance, nil
	} else {
		st.logger.Errorw("storage.Balance", "err", err.Error())
		return nil, domain.ErrServerInternal
	}
}

func (st *storage) UpdateBalanceByOrder(ctx context.Context, balance *domain.UserBalance, orderData *domain.OrderData) error {
	if balance == nil {
		st.logger.Errorw("storage.UpdateBalanceByOrder", "err", "balance is nil")
		return fmt.Errorf("%w: balance is nil", domain.ErrServerInternal)
	}

	if orderData == nil {
		st.logger.Errorw("storage.UpdateBalanceByOrder", "err", "orderData is nil")
		return fmt.Errorf("%w: orderData is nil", domain.ErrServerInternal)
	}

	tx, err := st.pPool.Begin(ctx)

	if err != nil {
		st.logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	var orderNum string
	err = tx.QueryRow(ctx,
		`update orderData set status = $1 where number = $2 and status<> $1 returning number`,
		orderData.Status,
		orderData.Number,
	).Scan(&orderNum)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("storage.UpdateBalanceByOrder", "status", "not found")
			return domain.ErrNotFound
		}
		st.logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	var balanceId int
	err = tx.QueryRow(ctx,
		`update balance set current = $1, withdrawn = $2 ,release = release+1 where userId=$3 and release=$4 returning balanceId`,
		balance.Current,
		balance.Withdrawn,
		balance.UserID,
		balance.Release).Scan(&balanceId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("storage.UpdateBalanceByOrder", "status", "not found")
			return domain.ErrBalanceChanged
		}
		st.logger.Errorw("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	if err = tx.Commit(ctx); err != nil {
		st.logger.Infow("storage.UpdateBalanceByOrder", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}

func (st *storage) UpdateBalanceByWithdraw(ctx context.Context, balance *domain.UserBalance, withdraw *domain.WithdrawalData) error {
	if balance == nil {
		st.logger.Errorw("storage.UpdateBalanceByWithdraw", "err", "balance is nil")
		return fmt.Errorf("%w: balance is nil", domain.ErrServerInternal)
	}

	if withdraw == nil {
		st.logger.Errorw("storage.UpdateBalanceByWithdraw", "err", "withdraw is nil")
		return fmt.Errorf("%w: withdraw is nil", domain.ErrServerInternal)
	}

	tx, err := st.pPool.Begin(ctx)

	if err != nil {
		st.logger.Errorw("storage.UpdateBatch", "err", err.Error())
		return domain.ErrServerInternal
	}

	defer tx.Rollback(ctx)

	tx.Exec(ctx,
		`insert into withdrawal(balancerId, number, sum, processed_at) values($1, $2, $3, $4)`,
		balance.BalanceId,
		withdraw.Order,
		withdraw.Sum,
		time.Time(withdraw.ProcessedAt),
	)

	var balanceId int
	err = tx.QueryRow(ctx,
		`update balance set current = $1, withdrawn = $2 ,release = release+1 where userId=$3 and release=$4 returning balanceId`,
		balance.Current,
		balance.Withdrawn,
		balance.UserID,
		balance.Release).Scan(&balanceId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("storage.UpdateBalanceByWithdraw", "status", "not found")
			return domain.ErrBalanceChanged
		}
		st.logger.Errorw("storage.UpdateBalanceByWithdraw", "err", err.Error())
		return domain.ErrServerInternal
	}

	if err = tx.Commit(ctx); err != nil {
		st.logger.Infow("storage.UpdateBalanceByWithdraw", "err", err.Error())
		return domain.ErrServerInternal
	}

	return nil
}

func (st *storage) Withdrawals(ctx context.Context, userID int) ([]domain.WithdrawalData, error) {
	var withdrawals []domain.WithdrawalData

	rows, err := st.pPool.Query(ctx,
		`select w.number, w.sum, w.processed_at from withdrawal w 
		inner join balance b on b.balanceId = w.balanceId where b.userId=1`,
		userID,
	)

	if err != nil {
		st.logger.Infow("storage.Withdrawals", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	defer rows.Close()

	for rows.Next() {
		var withdrawal domain.WithdrawalData
		var processed_at time.Time
		err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &processed_at)
		if err != nil {
			st.logger.Infow("storage.Withdrawals", "err", err.Error())
			return nil, domain.ErrServerInternal
		}
		withdrawal.ProcessedAt = domain.RFC3339Time(processed_at)
		withdrawals = append(withdrawals, withdrawal)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("storage.Withdrawals", "status", "not found")
			return nil, domain.ErrNotFound
		}
		st.logger.Infow("storage.Withdrawals", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return withdrawals, nil
}

package withdrawals

import (
	"context"
	"fmt"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/cache"
	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	dbinterface "github.com/Raime-34/gophermart.git/internal/repositories/db_interface"
	"github.com/Raime-34/gophermart.git/internal/utils"
	"github.com/google/uuid"
)

type WithdrawalsRepo struct {
	db                dbinterface.DbIface
	cachedWithdrawals *cache.Cache[*dto.WithdrawInfo]
	mu                sync.RWMutex
}

func NewWithdrawalsRepo(pool dbinterface.DbIface) *WithdrawalsRepo {
	return &WithdrawalsRepo{
		db:                pool,
		cachedWithdrawals: cache.NewCache[*dto.WithdrawInfo](),
	}
}

func (r *WithdrawalsRepo) RegisterWithdraw(ctx context.Context, req dto.WithdrawRequest) error {
	userId := ctx.Value(consts.UserIdKey)
	switch t := userId.(type) {
	case string:
	case uuid.UUID:
	default:
		return fmt.Errorf("Invalid userId type: %T", t)
	}

	userIdStr, _ := userId.(string)
	withdrawlInfo := dto.NewWithdrawInfo(req)

	_, err := r.db.Exec(ctx, insertWithdrawlQuery(), withdrawlInfo.Order, userIdStr, withdrawlInfo.Sum, withdrawlInfo.ProcessedAt)
	if err == nil {
		r.cachedWithdrawals.Set(utils.GetOrderInfoKey(userIdStr, withdrawlInfo.Order), withdrawlInfo)
	}

	return err
}

func (r *WithdrawalsRepo) GetWithdraws(ctx context.Context) ([]*dto.WithdrawInfo, error) {
	userId := ctx.Value(consts.UserIdKey)
	switch t := userId.(type) {
	case string:
	case uuid.UUID:
	default:
		return nil, fmt.Errorf("Invalid userId type: %T", t)
	}

	userIdStr, _ := userId.(string)
	withdrawls := r.cachedWithdrawals.GetByPrefix(utils.GetOrderInfoKeyPrefix(userIdStr))
	if len(withdrawls) > 0 {
		return withdrawls, nil
	}

	row, err := r.db.Query(ctx, getWithdrawalsQuery(), userIdStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to get withdrawls from db: %v", err)
	}
	defer row.Close()

	withdrawls = []*dto.WithdrawInfo{}
	for row.Next() {
		var withdrawl dto.WithdrawInfo
		if err := row.Scan(
			&withdrawl.Order,
			&withdrawl.Sum,
			&withdrawl.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("GetWithdraws: %w", err)
		}

		withdrawls = append(withdrawls, &withdrawl)
	}

	return withdrawls, nil
}

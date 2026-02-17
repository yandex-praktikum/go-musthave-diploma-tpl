package orders

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

type OrderRepo struct {
	db           dbinterface.DbIface
	cachedOrders *cache.Cache[*dto.OrderInfo]
	mu           sync.RWMutex
}

func NewOrderRepo(pool dbinterface.DbIface) *OrderRepo {
	return &OrderRepo{
		db:           pool,
		cachedOrders: cache.NewCache[*dto.OrderInfo](),
	}
}

func (r *OrderRepo) RegisterOrder(ctx context.Context, orderNumber string) error {
	userId := ctx.Value(consts.UserIdKey)
	switch t := userId.(type) {
	case string:
	case uuid.UUID:
	default:
		return fmt.Errorf("Invalid userId type: %T", t)
	}

	userIdStr, _ := userId.(string)
	orderInfo := dto.NewOrderInfo(orderNumber)

	_, err := r.db.Exec(ctx, insertOrderQuery(), orderInfo.Number, userId, orderInfo.Status, orderInfo.Accrual)
	if err == nil {
		r.cachedOrders.Set(utils.GetOrderInfoKey(userIdStr, orderInfo.Number), orderInfo)
	}
	return err
}

func (r *OrderRepo) UpdateOrder(ctx context.Context, newOrderState dto.AccrualCalculatorDTO) error {
	_, err := r.db.Exec(ctx, updateOrderQuery(), newOrderState.Status, newOrderState.Accrual, newOrderState.Order)
	return err
}

func (r *OrderRepo) GetOrders(ctx context.Context) ([]*dto.OrderInfo, error) {
	userId := ctx.Value(consts.UserIdKey)
	switch t := userId.(type) {
	case string:
	case uuid.UUID:
	default:
		return nil, fmt.Errorf("Invalid userId type: %v", t)
	}

	userIdStr, _ := userId.(string)
	orders := r.cachedOrders.GetByPrefix(utils.GetOrderInfoKeyPrefix(userIdStr))
	if len(orders) > 0 {
		return orders, nil
	}

	row, err := r.db.Query(ctx, getOrdersQuery(), userIdStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to get orders from db: %w", err)
	}
	defer row.Close()

	orders = []*dto.OrderInfo{}
	for row.Next() {
		var order dto.OrderInfo

		if err := row.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("GetOrders: %w", err)
		}

		orders = append(orders, &order)
	}

	for _, order := range orders {
		r.cachedOrders.Set(utils.GetOrderInfoKey(userIdStr, order.Number), order)
	}

	return orders, nil
}

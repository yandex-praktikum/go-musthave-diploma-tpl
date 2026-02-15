package orders

import (
	"context"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	dbinterface "github.com/Raime-34/gophermart.git/internal/repositories/db_interface"
)

type OrderRepo struct {
	db           dbinterface.DbIface
	cachedOrders map[string]*dto.OrderInfo
	mu           sync.RWMutex
}

func NewOrderRepo(pool dbinterface.DbIface) *OrderRepo {
	return &OrderRepo{
		db:           pool,
		cachedOrders: make(map[string]*dto.OrderInfo),
	}
}

func (r *OrderRepo) RegisterOrder(ctx context.Context, orderNumber string) error {
	_, err := r.db.Exec(ctx, insertOrderQuery(), orderNumber, ctx.Value(consts.UserIdKey), consts.REGISTERED, 0)
	return err
}

func (r *OrderRepo) UpdateOrder(ctx context.Context, newOrderState dto.AccrualCalculatorDTO) error {
	_, err := r.db.Exec(ctx, updateOrderQuery(), newOrderState.Status, newOrderState.Accrual, newOrderState.Order)
	return err
}

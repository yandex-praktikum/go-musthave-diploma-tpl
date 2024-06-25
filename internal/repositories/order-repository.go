package repositories

import (
	"context"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"gorm.io/gorm"
)

var orderRepository *OrderRepository

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	orderRepository = &OrderRepository{
		db: db,
	}

	return orderRepository
}

func (r *OrderRepository) Migrate(ctx context.Context) error {
	m := &entities.Order{}
	return r.db.WithContext(ctx).AutoMigrate(&m)
}

func (r *OrderRepository) Create(number string, userID uint) (*entities.Order, error) {
	order := &entities.Order{
		Number: number,
		UserID: userID,
		Status: entities.OrderStatusNew,
	}

	tx := r.db.Model(&entities.Order{}).
		Create(&order)
	err := tx.Error
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) FindByNumber(number string) (*entities.Order, error) {
	order := &entities.Order{}

	query := r.db.Where("orders.number = ?", number)

	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrdersByUserID(userID uint) ([]*models.GetOrdersResponse, error) {
	var orders []*models.GetOrdersResponse

	query := r.db.
		Table("orders").
		Select(`
			orders.number     as number,
			orders.status     as status,
			orders.accrual    as accrual,
			orders.created_at as uploaded_at
		`).
		Order("orders.created_at").
		Where("orders.user_id = ?", userID)

	if err := query.Scan(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetOrdersForProcess() ([]*entities.Order, error) {
	var orders []*entities.Order

	query := r.db.
		Table("orders").
		Select(`
			orders.number     as id,
			orders.created_at as created_at,
			orders.updated_at as updated_at,
			orders.number     as number,
			orders.status     as status,
			orders.accrual    as accrual
		`).
		Order("orders.created_at").
		Where("orders.status in ?", []entities.OrderStatus{entities.OrderStatusNew, entities.OrderStatusProcessing}).
		Where("orders.deleted_at is null").
		Limit(100)

	if err := query.Scan(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderByAccrualOrder(accrualOrder *models.AccrualOrderResponse) error {
	return r.db.Table("orders").Where("orders.number = ?", accrualOrder.Order).Updates(map[string]interface{}{
		"status":  accrualOrder.Status,
		"accrual": accrualOrder.Accrual,
	}).Error
}

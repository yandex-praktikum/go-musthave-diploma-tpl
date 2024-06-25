package repositories

import (
	"context"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"gorm.io/gorm"
	"time"
)

var operationRepository *OperationRepository

type OperationRepository struct {
	db *gorm.DB
}

func NewOperationRepository(db *gorm.DB) *OperationRepository {
	operationRepository = &OperationRepository{
		db: db,
	}

	return operationRepository
}

func (r *OperationRepository) Migrate(ctx context.Context) error {
	m := &entities.Operation{}
	return r.db.WithContext(ctx).AutoMigrate(&m)
}

func (r *OperationRepository) GetWithdrawnByAccountID(accountID uint) (float32, error) {
	var withdrawn float32

	query := r.db.
		Table("operations").
		Select(`
			coalesce(sum(operations.sum), 0) as withdrawn
		`).
		Where("operations.sender_account_id = ?", accountID).
		Where("operations.type = ?", entities.OperationTypeWithdraw).
		Where("operations.deleted_at is null").
		Where("operations.processed_at is not null")

	if err := query.Scan(&withdrawn).Error; err != nil {
		return 0, err
	}

	return withdrawn, nil
}

func (r *OperationRepository) CreateWithdrawn(accountID uint, orderNumber string, sum float32) error {
	systemWithdrawnAccount, err := accountRepository.GetSystemWithdrawnAccountID()

	if err != nil {
		return err
	}
	if systemWithdrawnAccount == 0 {
		return errors.New("cannot create withdrawn")
	}

	now := time.Now()

	tx := r.db.Begin()
	err = tx.Table("operations").
		Create(map[string]interface{}{
			"created_at":           now,
			"updated_at":           now,
			"processed_at":         now,
			"type":                 entities.OperationTypeWithdraw,
			"order_number":         orderNumber,
			"sum":                  sum,
			"sender_account_id":    accountID,
			"recipient_account_id": systemWithdrawnAccount,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = accountRepository.Transaction(accountID, systemWithdrawnAccount, sum)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (r *OperationRepository) CreateAccrual(accountID uint, orderNumber string, sum float32) error {
	systemWithdrawnAccount, err := accountRepository.GetSystemWithdrawnAccountID()

	if err != nil {
		return err
	}
	if systemWithdrawnAccount == 0 {
		return errors.New("cannot create accrual")
	}

	now := time.Now()

	tx := r.db.Begin()
	err = tx.Table("operations").
		Create(map[string]interface{}{
			"created_at":           now,
			"updated_at":           now,
			"processed_at":         now,
			"type":                 entities.OperationTypeAccrual,
			"order_number":         orderNumber,
			"sum":                  sum,
			"sender_account_id":    systemWithdrawnAccount,
			"recipient_account_id": accountID,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = accountRepository.Transaction(systemWithdrawnAccount, accountID, sum)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (r *OperationRepository) GetWithdrawalsByAccountID(accountID uint) ([]models.GetWithdrawalsResponse, error) {
	var operations []models.GetWithdrawalsResponse

	err := r.db.Table("operations").
		Select(`
			operations.order_number as order,
			operations.sum as sum,
			operations.processed_at as processed_at
		`).
		Where("operations.sender_account_id = ?", accountID).Scan(&operations).Error
	if err != nil {
		return nil, err
	}

	return operations, nil
}

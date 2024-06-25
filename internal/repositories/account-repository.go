package repositories

import (
	"context"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"gorm.io/gorm"
)

var accountRepository *AccountRepository

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	accountRepository = &AccountRepository{
		db: db,
	}

	return accountRepository
}

func (r *AccountRepository) Migrate(ctx context.Context) error {
	m := &entities.Account{}
	return r.db.WithContext(ctx).AutoMigrate(&m)
}

func (r *AccountRepository) Create(userID uint, accountType entities.AccountType) (*entities.Account, error) {
	account := &entities.Account{
		Type:   accountType,
		UserID: userID,
	}

	tx := r.db.Model(&entities.Account{}).
		Create(&account)
	err := tx.Error
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *AccountRepository) FindByUserID(userID uint, accountType entities.AccountType) (*entities.Account, error) {
	account := &entities.Account{}

	query := r.db.
		Where("accounts.type = ?", accountType).
		Where("accounts.user_id = ?", userID)

	if err := query.First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return account, nil
}

func (r *AccountRepository) GetSystemWithdrawnAccountID() (uint, error) {
	var accountID uint

	query := r.db.
		Table("accounts").
		Select(`
			coalesce(accounts.id, 0) as account_id
		`).
		Where("accounts.type = ?", entities.AccountTypeSystemWithdraw).
		Where("accounts.deleted_at is null")

	if err := query.Scan(&accountID).Error; err != nil {
		return 0, err
	}

	return accountID, nil
}

func (r *AccountRepository) Transaction(senderAccountID uint, recipientAccountID uint, sum float32) error {
	var senderAccount entities.Account

	err := r.db.Table("accounts").Where("accounts.id = ?", senderAccountID).Find(&senderAccount).Error
	if err != nil {
		return err
	}

	var recipientAccount entities.Account
	err = r.db.Table("accounts").Where("accounts.id = ?", recipientAccountID).Find(&recipientAccount).Error
	if err != nil {
		return err
	}

	err = r.db.Table("accounts").Where("accounts.id = ?", senderAccountID).Update("sum", senderAccount.Sum-sum).Error
	if err != nil {
		return err
	}

	err = r.db.Table("accounts").Where("accounts.id = ?", recipientAccountID).Update("sum", recipientAccount.Sum+sum).Error
	if err != nil {
		return err
	}

	return nil
}

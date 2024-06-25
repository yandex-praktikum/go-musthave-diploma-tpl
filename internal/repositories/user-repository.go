package repositories

import (
	"context"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var userRepository *UserRepository

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	userRepository = &UserRepository{
		db: db,
	}

	return userRepository
}

func (r *UserRepository) Migrate(ctx context.Context) error {
	m := &entities.User{}
	return r.db.WithContext(ctx).AutoMigrate(&m)
}

func (r *UserRepository) Create(userRegister models.UserRegisterRequest) (*models.UserInfoResponse, error) {
	passwordHash, err := r.GeneratePasswordHash(userRegister.Password)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Login:    userRegister.Login,
		Password: string(passwordHash),
	}

	tx := r.db.Begin()

	query := tx.Model(&entities.User{}).
		Create(&user)
	err = query.Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = accountRepository.Create(user.ID, entities.AccountTypeFree)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = accountRepository.Create(user.ID, entities.AccountTypeBonus)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &models.UserInfoResponse{
		ID:         user.ID,
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		MiddleName: user.MiddleName,
		Login:      user.Login,
		Email:      user.Email,
	}, nil
}

func (r *UserRepository) Find(id uint) (*models.UserInfoResponse, error) {
	userModel := &models.UserInfoResponse{}
	if err := r.db.
		Select(`
		    users.id                                as id,
		    users.first_name                        as first_name,
		    users.middle_name                       as middle_name,
		    users.last_name                         as last_name,
		    users.login                             as login,
		    users.email                             as email`).
		Table("users").
		Where("users.id = ?", id).
		Where("users.deleted_at is null").
		Scan(&userModel).
		Error; err != nil {
		return nil, err
	}

	return userModel, nil
}

func (r *UserRepository) FindBy(filter models.UserSearchFilter) (*entities.User, error) {
	user := &entities.User{}

	query := r.db

	if filter.Login != "" {
		query = query.Where("\"users\".\"login\" = ?", filter.Login)
	}

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GeneratePasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 8)
}

package services

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/utils"
	"github.com/google/uuid"
)

type AuthStore interface {
	SelectUserByUsername(ctx context.Context, username string) (*models.User, error)
	InsertUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

type AuthService struct {
	SecretKey string
	authStore AuthStore
}

func NewAuthService(SecretKey string, authStore AuthStore) *AuthService {
	return &AuthService{SecretKey: SecretKey, authStore: authStore}
}

func (as *AuthService) CreateUser(ctx context.Context, username string, password string) (*models.User, error) {
	u, _ := as.authStore.SelectUserByUsername(ctx, username)
	if u != nil {
		return nil, fmt.Errorf("there's a registered user with this username address")
	}
	hashPassword := utils.HashPassword(password, as.SecretKey)
	user := &models.User{
		ID:       uuid.New(),
		Username: username,
		Password: hashPassword,
	}
	err := as.authStore.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (as *AuthService) GetUser(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := as.authStore.SelectUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	hashPassword := utils.HashPassword(password, as.SecretKey)
	if hashPassword != user.Password {
		return nil, fmt.Errorf("invalid password")
	}
	return user, nil
}

func (as *AuthService) ChangePassword(ctx context.Context, username string, password string) error {
	user, err := as.authStore.SelectUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	user.Password = utils.HashPassword(password, as.SecretKey)
	user.RefreshToken = ""
	err = as.authStore.UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

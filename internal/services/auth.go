package services

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	SecretKey string
	authStore AuthStore
}

func NewAuthService(SecretKey string, authStore AuthStore) *AuthService {
	return &AuthService{SecretKey: SecretKey, authStore: authStore}
}

func (as *AuthService) CreateUser(ctx context.Context, email string, password string) (*models.User, error) {
	err := utils.ValidateEmail(email)
	if err != nil {
		return nil, err
	}
	u, _ := as.authStore.SelectUserByEmail(ctx, email)
	if u != nil {
		return nil, fmt.Errorf("there's a registered user with this e-mail address")
	}
	hashPassword := utils.HashPassword(password, as.SecretKey)
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: hashPassword,
	}
	err = as.authStore.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (as *AuthService) GetUser(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := as.authStore.SelectUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	hashPassword := utils.HashPassword(password, as.SecretKey)
	if hashPassword != user.Password {
		return nil, fmt.Errorf("invalid password")
	}
	return user, nil
}

func (as *AuthService) ChangePassword(ctx context.Context, email string, password string) error {
	user, err := as.authStore.SelectUserByEmail(ctx, email)
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

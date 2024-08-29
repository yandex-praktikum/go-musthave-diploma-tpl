package services

import (
	"context"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

type UserStore interface {
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	SelectUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type UserService struct {
	userStore UserStore
}

func NewUserService(userStore UserStore) *UserService {
	return &UserService{userStore: userStore}
}

func (us *UserService) UpdateUser(
	ctx context.Context,
	user *models.User,
	name *string,
	age *uint8,
) error {
	if name != nil {
		user.Name = *name
	}
	if age != nil {
		user.Age = *age
	}
	return us.userStore.UpdateUser(ctx, user)
}

func (us *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return us.userStore.DeleteUser(ctx, userID)
}

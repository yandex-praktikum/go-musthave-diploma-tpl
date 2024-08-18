package services

import (
	"context"
	"fmt"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/utils"
	"github.com/google/uuid"
)

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
	email *string,
) error {
	if email != nil {
		email := *email
		if email != user.Email {
			err := utils.ValidateEmail(email)
			if err != nil {
				return err
			}
			u, _ := us.userStore.SelectUserByEmail(ctx, email)
			if u != nil {
				return fmt.Errorf("there's a registered user with this e-mail address")
			}
			user.Email = email
			user.EmailConfirmed = false
		}
	}
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

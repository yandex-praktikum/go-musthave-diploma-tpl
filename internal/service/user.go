package service

import (
	"github.com/sub3er0/gophermart/internal/models"
	"github.com/sub3er0/gophermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserRepository *repository.UserRepository
}

func (us *UserService) IsUserExist(username string) int {
	return us.UserRepository.IsUserExists(username)
}

func (us *UserService) RegisterUser(user models.User) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return user, err
	}

	user.Password = string(hashedPassword)

	err = us.UserRepository.CreateUser(user)

	if err != nil {
		return user, err
	}

	err = us.UserRepository.CreateUserBalance(user)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (us *UserService) AuthenticateUser(username, password string) (models.User, error) {
	user, err := us.UserRepository.GetUserByUsername(username)

	if err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, err
	}

	return user, nil
}

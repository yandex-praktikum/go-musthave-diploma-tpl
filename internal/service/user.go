package service

import (
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/models"
	"gophermart/internal/repository"
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

	user.ID, err = us.UserRepository.CreateUser(user)

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

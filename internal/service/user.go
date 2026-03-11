package service

import (
	"context"
	"errors"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService — бизнес-логика пользователей (регистрация, логин).
type UserService struct {
	repo repository.UserRepository
}

// NewUserService создаёт сервис пользователей.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register создаёт пользователя. При занятом логине возвращает *repository.ErrDuplicateLogin.
func (s *UserService) Register(ctx context.Context, login, password string) (*models.User, error) {
	if login == "" || password == "" {
		return nil, &ErrValidation{Msg: "login and password required"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.Create(ctx, login, string(hash))
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Login проверяет логин/пароль и возвращает пользователя. При неверных данных — *ErrInvalidCredentials или *repository.ErrUserNotFound.
func (s *UserService) Login(ctx context.Context, login, password string) (*models.User, error) {
	if login == "" || password == "" {
		return nil, &ErrValidation{Msg: "login and password required"}
	}
	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		var notFound *repository.ErrUserNotFound
		if errors.As(err, &notFound) {
			return nil, &ErrInvalidCredentials{Login: login}
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, &ErrInvalidCredentials{Login: login}
	}
	return user, nil
}

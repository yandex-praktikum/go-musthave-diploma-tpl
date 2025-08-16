package service

import (
	"context"
	"errors"
	"strings"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/auth"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, login, passwordHash string) error
	IsLoginExist(ctx context.Context, login string) (bool, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type UserService struct {
	UserRepo UserRepo
}

var (
	ErrUserExists      = errors.New("логин уже занят")
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrInvalidPassword = errors.New("неверная пара логин/пароль")
)

func NewUserService(repo UserRepo) *UserService {
	return &UserService{UserRepo: repo}
}

func (s *UserService) Register(ctx context.Context, req models.RegisterRequest) (string, error) {
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		return "", errors.New("неверный формат запроса")
	}
	exists, err := s.UserRepo.IsLoginExist(ctx, req.Login)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := s.UserRepo.CreateUser(ctx, req.Login, string(hash)); err != nil {
		return "", err
	}
	token, err := auth.GenerateJWT(req.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) Login(ctx context.Context, req models.RegisterRequest) (string, error) {
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		return "", errors.New("неверный формат запроса")
	}
	user, err := s.UserRepo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		return "", ErrUserNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return "", ErrInvalidPassword
	}
	token, err := auth.GenerateJWT(req.Login)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.UserRepo.GetUserByLogin(ctx, login)
}

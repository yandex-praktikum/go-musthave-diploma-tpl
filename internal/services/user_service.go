// Package services предоставляет бизнес-логику приложения.
package services

import (
	"context"
	"errors"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/storage"
)

// Ошибки сервиса пользователей
var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrLoginAlreadyExists = errors.New("login already exists")
)

// UserService предоставляет методы для работы с пользователями
type UserService struct {
	userRepo     *storage.UserRepository
	tokenManager *auth.TokenManager
}

// NewUserService создаёт новый сервис пользователей
func NewUserService(userRepo *storage.UserRepository, tokenManager *auth.TokenManager) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

// Register регистрирует нового пользователя и возвращает токен
func (s *UserService) Register(ctx context.Context, login, password string) (string, error) {
	// Хешируем пароль
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Создаём пользователя
	user, err := s.userRepo.Create(ctx, login, []byte(hash))
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return "", ErrLoginAlreadyExists
		}
		return "", err
	}

	// Генерируем токен
	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login аутентифицирует пользователя и возвращает токен
func (s *UserService) Login(ctx context.Context, login, password string) (string, error) {
	// Находим пользователя
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Проверяем пароль
	// HashPassword возвращает строку, а у нас []byte в БД
	if !auth.VerifyPassword(password, string(user.PasswordHash)) {
		return "", ErrInvalidCredentials
	}

	// Генерируем токен
	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserByID находит пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*storage.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

package service

import (
	"errors"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService — сервис регистрации и аутентификации пользователей.
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService создаёт новый AuthService с указанным репозиторием пользователей.
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register регистрирует нового пользователя.
//
// Проверяет уникальность логина, хеширует пароль через bcrypt и сохраняет
// пользователя в БД. Возвращает domain.ErrLoginAlreadyExists, если логин занят.
func (s *AuthService) Register(login, password string) (*domain.User, error) {
	// Проверяем, существует ли пользователь
	exists, err := s.userRepo.Exists(login)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrLoginAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Login:     login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login аутентифицирует пользователя по паре логин/пароль.
//
// Возвращает domain.ErrInvalidCredentials, если пользователь не найден
// или пароль не совпадает.
func (s *AuthService) Login(login, password string) (*domain.User, error) {
	user, err := s.userRepo.FindByLogin(login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

// Package service реализует бизнес-логику сервиса
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/anon-d/gophermarket/internal/domain"
	"github.com/anon-d/gophermarket/internal/http/helper"
	"github.com/anon-d/gophermarket/internal/repository"
	"github.com/anon-d/gophermarket/internal/repository/postgres"
	"github.com/anon-d/gophermarket/pkg/jwt"
	"github.com/anon-d/gophermarket/pkg/luhn"
)

// Repository интерфейс хранилища данных, используемый сервисом
//
//go:generate go run go.uber.org/mock/mockgen -destination=mock_repository_test.go -package=service github.com/anon-d/gophermarket/internal/http/service Repository
type Repository interface {
	CreateUser(ctx context.Context, user *repository.User) error
	GetUserByLogin(ctx context.Context, login string) (*repository.User, error)
	CreateOrder(ctx context.Context, order *repository.Order) error
	GetOrdersByUserID(ctx context.Context, userID string) ([]repository.Order, error)
	GetBalance(ctx context.Context, userID string) (*repository.Balance, error)
	Withdraw(ctx context.Context, withdrawal *repository.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) ([]repository.Withdrawal, error)
}

var (
	// ErrUserExists ошибка - пользователь уже существует
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials ошибка - неверные учётные данные
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrOrderExists ошибка - заказ уже загружен этим пользователем
	ErrOrderExists = errors.New("order already exists")
	// ErrOrderExistsByAnotherUser ошибка - заказ загружен другим пользователем
	ErrOrderExistsByAnotherUser = errors.New("order exists by another user")
	// ErrInvalidOrderNumber ошибка - неверный формат номера заказа
	ErrInvalidOrderNumber = errors.New("invalid order number")
	// ErrInsufficientFunds ошибка - недостаточно средств
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// GopherService сервис бизнес-логики
type GopherService struct {
	namespace     string
	repo          Repository
	logger        *zap.Logger
	jwtSecret     string
	tokenDuration time.Duration
}

func NewGopherService(namespace string, repo Repository, logger *zap.Logger, jwtSecret string) *GopherService {
	return &GopherService{
		namespace:     namespace,
		repo:          repo,
		logger:        logger,
		jwtSecret:     jwtSecret,
		tokenDuration: 24 * time.Hour,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *GopherService) RegisterUser(ctx context.Context, login, password string) (string, error) {
	op := "service:RegisterUser"

	space := uuid.MustParse(s.namespace)
	uid := uuid.NewSHA1(space, []byte(login))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Ошибка создания хеша пароля", zap.String("op", op), zap.Error(err))
		return "", err
	}

	userDomain := domain.User{
		ID:       uid.String(),
		Login:    login,
		PassHash: string(passHash),
	}

	err = s.repo.CreateUser(ctx, helper.ToRepositoryUser(&userDomain))
	if err != nil {
		if errors.Is(err, postgres.ErrUserExists) {
			return "", ErrUserExists
		}
		s.logger.Error("Ошибка создания пользователя в БД", zap.String("op", op), zap.Error(err))
		return "", err
	}

	// Генерируем токен для автоматической аутентификации
	token, err := jwt.NewToken(uid.String(), s.tokenDuration, s.jwtSecret)
	if err != nil {
		s.logger.Error("Ошибка создания токена", zap.String("op", op), zap.Error(err))
		return "", err
	}

	return token, nil
}

// LoginUser аутентифицирует пользователя
func (s *GopherService) LoginUser(ctx context.Context, login, password string) (string, error) {
	op := "service:LoginUser"

	userRepository, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		s.logger.Error("Ошибка получения пользователя", zap.String("op", op), zap.Error(err))
		return "", err
	}

	user := helper.ToDomainUser(userRepository)

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := jwt.NewToken(user.ID, s.tokenDuration, s.jwtSecret)
	if err != nil {
		s.logger.Error("Ошибка создания токена", zap.String("op", op), zap.Error(err))
		return "", err
	}

	return token, nil
}

// CreateOrder создаёт новый заказ
func (s *GopherService) CreateOrder(ctx context.Context, userID, orderNumber string) error {
	op := "service:CreateOrder"

	// Проверяем номер заказа по алгоритму Луна
	if !luhn.Valid(orderNumber) {
		return ErrInvalidOrderNumber
	}

	order := &domain.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     domain.OrderStatusNew,
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	err := s.repo.CreateOrder(ctx, helper.ToRepositoryOrder(order))
	if err != nil {
		if errors.Is(err, postgres.ErrOrderExists) {
			return ErrOrderExists
		}
		if errors.Is(err, postgres.ErrOrderExistsByAnotherUser) {
			return ErrOrderExistsByAnotherUser
		}
		s.logger.Error("Ошибка создания заказа", zap.String("op", op), zap.Error(err))
		return err
	}

	return nil
}

// GetOrders получает список заказов пользователя
func (s *GopherService) GetOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	op := "service:GetOrders"

	orders, err := s.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Ошибка получения заказов", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return helper.ToDomainOrders(orders), nil
}

// GetBalance получает баланс пользователя
func (s *GopherService) GetBalance(ctx context.Context, userID string) (*domain.Balance, error) {
	op := "service:GetBalance"

	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Ошибка получения баланса", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return helper.ToDomainBalance(balance), nil
}

// Withdraw списывает баллы со счёта пользователя
func (s *GopherService) Withdraw(ctx context.Context, userID, orderNumber string, sum float64) error {
	op := "service:Withdraw"

	// Проверяем номер заказа по алгоритму Луна
	if !luhn.Valid(orderNumber) {
		return ErrInvalidOrderNumber
	}

	withdrawal := &domain.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}

	err := s.repo.Withdraw(ctx, helper.ToRepositoryWithdrawal(withdrawal))
	if err != nil {
		if errors.Is(err, postgres.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}
		s.logger.Error("Ошибка списания баллов", zap.String("op", op), zap.Error(err))
		return err
	}

	return nil
}

// GetWithdrawals получает список списаний пользователя
func (s *GopherService) GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error) {
	op := "service:GetWithdrawals"

	withdrawals, err := s.repo.GetWithdrawals(ctx, userID)
	if err != nil {
		s.logger.Error("Ошибка получения списаний", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return helper.ToDomainWithdrawals(withdrawals), nil
}

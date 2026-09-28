// Package service contains Gophermart business logic.
package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/true18/go-musthave-diploma-tpl/internal/luhn"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
)

// Domain errors returned by Service.
var (
	ErrUserExists              = repository.ErrUserExists
	ErrOrderExistsSameUser     = repository.ErrOrderExistsSameUser
	ErrOrderExistsAnotherUser  = repository.ErrOrderExistsAnotherUser
	ErrInsufficientFunds       = repository.ErrInsufficientFunds
	ErrNotFound                = repository.ErrNotFound
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidOrderFormat      = errors.New("invalid order format")
	ErrInvalidOrderNumber      = errors.New("invalid order number")
	ErrInvalidWithdrawalAmount = errors.New("invalid withdrawal amount")
)

// Repository is the storage contract used by Service and Worker.
type Repository interface {
	CreateUser(ctx context.Context, login string, passwordHash []byte) (int64, error)
	FindUserByLogin(ctx context.Context, login string) (repository.User, error)
	CreateOrder(ctx context.Context, userID int64, number string) error
	ListUserOrders(ctx context.Context, userID int64) ([]repository.Order, error)
	ListPendingOrders(ctx context.Context, limit int) ([]repository.Order, error)
	UpdateOrder(ctx context.Context, number string, status string, accrual *money.Amount) error
	Balance(ctx context.Context, userID int64) (repository.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount money.Amount) error
	ListWithdrawals(ctx context.Context, userID int64) ([]repository.Withdrawal, error)
}

// Service provides user, order and balance operations.
type Service struct {
	repo Repository
}

// New creates a Service.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a user and returns its id.
func (s *Service) Register(ctx context.Context, login string, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	return s.repo.CreateUser(ctx, login, hash)
}

// Login verifies credentials and returns the user id.
func (s *Service) Login(ctx context.Context, login string, password string) (int64, error) {
	user, err := s.repo.FindUserByLogin(ctx, login)
	if errors.Is(err, repository.ErrNotFound) {
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return 0, ErrInvalidCredentials
	}
	return user.ID, nil
}

// UploadOrder validates and stores an order number for userID.
func (s *Service) UploadOrder(ctx context.Context, userID int64, number string) error {
	if !allDigits(number) {
		return ErrInvalidOrderFormat
	}
	if !luhn.Valid(number) {
		return ErrInvalidOrderNumber
	}
	return s.repo.CreateOrder(ctx, userID, number)
}

// ListOrders returns user orders from newest to oldest.
func (s *Service) ListOrders(ctx context.Context, userID int64) ([]repository.Order, error) {
	return s.repo.ListUserOrders(ctx, userID)
}

// Balance returns current and withdrawn user balance.
func (s *Service) Balance(ctx context.Context, userID int64) (repository.Balance, error) {
	return s.repo.Balance(ctx, userID)
}

// Withdraw atomically writes a withdrawal if the user has enough points.
func (s *Service) Withdraw(ctx context.Context, userID int64, orderNumber string, amount money.Amount) error {
	if !allDigits(orderNumber) || !luhn.Valid(orderNumber) {
		return ErrInvalidOrderNumber
	}
	if amount <= 0 {
		return ErrInvalidWithdrawalAmount
	}
	return s.repo.Withdraw(ctx, userID, orderNumber, amount)
}

// ListWithdrawals returns user withdrawals from newest to oldest.
func (s *Service) ListWithdrawals(ctx context.Context, userID int64) ([]repository.Withdrawal, error) {
	return s.repo.ListWithdrawals(ctx, userID)
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

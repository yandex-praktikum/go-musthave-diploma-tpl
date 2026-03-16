package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"gophermart/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (int64, error) {
	if login == "" || password == "" {
		return 0, ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	userID, err := s.userRepo.Create(ctx, login, string(hash))
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	return userID, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (int64, error) {
	if login == "" || password == "" {
		return 0, ErrInvalidInput
	}

	userID, passwordHash, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrUnauthorized
		}
		return 0, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return 0, ErrUnauthorized
	}

	return userID, nil
}

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	const duplicateKey = "duplicate key value violates unique constraint"
	return strings.Contains(err.Error(), duplicateKey)
}

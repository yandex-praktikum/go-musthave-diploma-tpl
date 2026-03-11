package repository

//go:generate mockgen -source=user.go -destination=mock/mock_user_repository.go -package=mock

import (
	"context"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
)

// UserRepository — работа с пользователями в БД (без бизнес-логики).
type UserRepository interface {
	Create(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

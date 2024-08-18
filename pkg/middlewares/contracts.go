package middlewares

import (
	"context"

	"github.com/eac0de/gophermart/internal/models"
	"github.com/google/uuid"
)

type UserStore interface {
	SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

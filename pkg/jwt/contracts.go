package jwt

import (
	"context"

	"github.com/google/uuid"
)

type RefreshTokenStore interface {
	SaveRefreshToken(ctx context.Context, RefreshToken string, userID *uuid.UUID) error
	GetRefreshToken(ctx context.Context, userID *uuid.UUID) string
}

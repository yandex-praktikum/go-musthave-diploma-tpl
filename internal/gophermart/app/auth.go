package app

import (
	"context"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewAuth() *auth {
	return nil
}

type auth struct {
	// salt []byte
}

func (a *auth) Register(ctx context.Context, regData *domain.RegistrationData) error {
	// TODO
	return nil
}

func (a *auth) Authentificate(ctx context.Context, userData *domain.LoginData) (domain.TokenString, error) {
	// TODO
	return "", nil
}

func (a *auth) Authorize(ctx context.Context, tokenString domain.TokenString) (*domain.AuthData, error) {
	// TODO
	return nil, nil
}

package user

import (
	"context"
	"time"

	"github.com/kdv2001/loyalty/internal/domain"
)

type authRepo interface {
	Register(ctx context.Context, login domain.Login) (domain.ID, error)
	Login(ctx context.Context, login domain.Login) (domain.Auth, error)
}

type SessionStore interface {
	SetSession(ctx context.Context, session domain.SessionInfo) (domain.SessionToken, error)
	GetSessions(ctx context.Context, token domain.SessionToken) (domain.SessionInfo, error)
}

type Repo interface {
	GetUser()
}

type Implementation struct {
	authRepo     authRepo
	SessionStore SessionStore
}

func NewImplementation(authRepo authRepo, SessionStore SessionStore) *Implementation {
	return &Implementation{
		authRepo:     authRepo,
		SessionStore: SessionStore,
	}
}

type Register struct {
	Login    string
	Password string
}

func (a *Implementation) RegisterAndLoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error) {
	_, err := a.RegisterUser(ctx, reg)
	if err != nil {
		return domain.SessionToken{}, err
	}

	token, err := a.LoginUser(ctx, reg)
	if err != nil {
		return domain.SessionToken{}, err
	}

	return token, nil
}

func (a *Implementation) LoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error) {
	auth, err := a.authRepo.Login(ctx, reg)
	if err != nil {
		return domain.SessionToken{}, err
	}

	token, err := a.SessionStore.SetSession(ctx, domain.SessionInfo{
		CreatedAt: time.Now().UTC(),
		UserID:    auth.UserID,
		Device:    "not used",
	})
	if err != nil {
		return domain.SessionToken{}, err
	}

	return token, nil
}

func (a *Implementation) AuthUser(ctx context.Context, token domain.SessionToken) (domain.SessionInfo, error) {
	session, err := a.SessionStore.GetSessions(ctx, token)
	if err != nil {
		return domain.SessionInfo{}, err
	}

	return session, nil
}

func (a *Implementation) RegisterUser(ctx context.Context, reg domain.Login) (domain.ID, error) {
	user, err := a.authRepo.Register(ctx, reg)
	if err != nil {
		return domain.ID{}, err
	}

	return user, nil
}

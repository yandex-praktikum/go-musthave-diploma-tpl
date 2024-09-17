package user_auth

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/korol8484/gofermart/internal/app/domain"
	"net/http"
)

type AuthUser interface {
	CreateUser(ctx context.Context, user *domain.User, password string) (*domain.User, error)
	Auth(ctx context.Context, login, password string) (*domain.User, error)
}

type AuthSession interface {
	CreateSession(w http.ResponseWriter, r *http.Request, id domain.UserId) error
}

// Handler -
// В задаче не учитывается, что делать, если сессия уже существует (аутентифицированый пользователь)
// поэтому такие нюансы тут не учитываем и считаем, что обращения на данные методы всегда выполняют поставленные задачи
type Handler struct {
	auth    AuthUser
	session AuthSession
}

func NewAuthHandler(auth AuthUser, session AuthSession) *Handler {
	return &Handler{
		auth:    auth,
		session: session,
	}
}

type request struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) RegisterRoutes() func(mux *chi.Mux) {
	return func(mux *chi.Mux) {
		mux.Post("/api/user/register", h.registerHandler)
		mux.Post("/api/user/login", h.authHandler)
	}
}

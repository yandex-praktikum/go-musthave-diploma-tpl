package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	http2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/http/utils/auth"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/service/user"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"time"
)

type JWTClient interface {
	BuildJWTString(userID int) (string, error)
	GetTokenExp() time.Duration
}

type Service interface {
	Ping(ctx context.Context) error
	Register(ctx context.Context, userRegister *entity.UserRegisterJSON) (*entity.UserDB, error)
}

type handler struct {
	logger    logging2.Logger
	service   Service
	jwtClient JWTClient
}

func NewHandler(logger logging2.Logger, updateService Service, jwtClient JWTClient) http2.Handler {
	return &handler{
		logger:    logger,
		service:   updateService,
		jwtClient: jwtClient,
	}
}

func (h handler) Register(router *chi.Mux) {
	router.Post("/api/user/register", h.userRegister)
}

// userRegister /api/user/register
func (h handler) userRegister(writer http.ResponseWriter, request *http.Request) {
	userRegister, err := decodeUserRegister(request.Body)
	if err != nil {
		h.logger.Error(err)
		// 400 — неверный формат запроса;
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Валидация body полей
	if len(userRegister.Password) == 0 || len(userRegister.Login) == 0 {
		// 400 — неверный формат запроса;
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	userDB, err := h.service.Register(request.Context(), userRegister)
	// 409 — логин уже занят;
	if errors.Is(err, user.ErrNotUniqueLogin) {
		writer.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		authToken, err := h.jwtClient.BuildJWTString(userDB.ID)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		auth.SetAuthCookie(writer, authToken, h.jwtClient.GetTokenExp())
		writer.WriteHeader(http.StatusOK)
		return
	}
}

func decodeUserRegister(body io.ReadCloser) (*entity.UserRegisterJSON, error) {
	var userRegister entity.UserRegisterJSON

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&userRegister)

	return &userRegister, err
}

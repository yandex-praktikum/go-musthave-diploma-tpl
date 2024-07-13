package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/domain/entity"
	http2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/service/user"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"time"
)

type JWTClient interface {
	BuildJWTString(userId int) (string, error)
	GetTokenExp() time.Duration
}

type Service interface {
	Login(ctx context.Context, userRegister *entity.UserLoginJSON) (*entity.UserDB, error)
}

type handler struct {
	logger        logging2.Logger
	updateService Service
	jwtClient     JWTClient
}

func NewHandler(logger logging2.Logger, updateService Service, jwtClient JWTClient) http2.Handler {
	return &handler{
		logger:        logger,
		updateService: updateService,
		jwtClient:     jwtClient,
	}
}

func (h handler) Register(router *chi.Mux) {
	router.Post("/api/user/login", h.userLogin)
}

// userRegister /api/user/login
func (h handler) userLogin(writer http.ResponseWriter, request *http.Request) {
	userLogin, err := decodeUserLogin(request.Body)
	if err != nil {
		h.logger.Error(err)
		// 400 — неверный формат запроса;
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Валидация body полей
	if len(userLogin.Password) == 0 || len(userLogin.Login) == 0 {
		// 400 — неверный формат запроса;
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	userDB, err := h.updateService.Login(request.Context(), userLogin)
	// 401 — неверная пара логин/пароль;
	if errors.Is(err, user.ErrInvalidLoginPasswordCombination) {
		writer.WriteHeader(http.StatusUnauthorized)
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

		setAuthCookie(writer, authToken, h.jwtClient.GetTokenExp())
		writer.WriteHeader(http.StatusOK)
		return
	}
}

func decodeUserLogin(body io.ReadCloser) (*entity.UserLoginJSON, error) {
	var userLogin entity.UserLoginJSON

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&userLogin)

	return &userLogin, err
}

func setAuthCookie(w http.ResponseWriter, authToken string, tokenExp time.Duration) {
	cookie := http.Cookie{
		Name:     "at", // accessToken
		Value:    authToken,
		Path:     "/",
		MaxAge:   int(tokenExp),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}

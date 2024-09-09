package register

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth"
	"net/http"
)

type Handler struct {
	ctx         context.Context
	authService *auth.ServiceAuth
	log         *logger.Logger
}

func NewHandlers(ctx context.Context, authService *auth.ServiceAuth, log *logger.Logger) *Handler {
	return &Handler{
		ctx:         ctx,
		authService: authService,
		log:         log,
	}
}

func (h *Handler) ServerHTTP(w http.ResponseWriter, r *http.Request) {

	// Считываем тело запроса и записываем в body
	body := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.log.Error("error register", "error:", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return

	}

	// проверяем есть ли пользователь и если нет
	if err = h.authService.RegisterUser(h.ctx, body.Login, body.Password); err != nil {
		if errors.Is(err, customerrors.ErrUserAlreadyExists) {
			h.log.Error("error register", "error:", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: err.Error()})
			w.WriteHeader(http.StatusConflict)
			w.Write(apiError)
			return
		}

		h.log.Error("error register", "error:", err)
		apoError, _ := json.Marshal(customerrors.APIError{Message: "you are not logged in"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apoError)
		return
	}

	// Аутентифицируем пользователя
	token, err := h.authService.AuthUser(h.ctx, body.Login, body.Password)
	if err != nil {
		h.log.Error("authentication error", "error:", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "authentication failed"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    token.AccessToken,
		HttpOnly: true,
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    token.RefreshToken,
		HttpOnly: true,
	}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	response, _ := json.Marshal(ResponseBody{Success: true})
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

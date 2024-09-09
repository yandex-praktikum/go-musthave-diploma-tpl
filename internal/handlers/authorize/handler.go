package authorize

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

func NewHandler(ctx context.Context, authService *auth.ServiceAuth, log *logger.Logger) *Handler {
	return &Handler{
		ctx:         ctx,
		authService: authService,
		log:         log,
	}
}

func (h *Handler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	body := RequestBody{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("error authorize", "error:", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	_, err := h.authService.AuthUser(h.ctx, body.Login, body.Password)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			h.log.Error("error authorize", "error:", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "cannot find user"})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(apiError)
			return
		}

		if errors.Is(err, customerrors.ErrIsTruePassword) {
			h.log.Error("error authorize", "error:", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect password"})
			w.WriteHeader(http.StatusForbidden)
			w.Write(apiError)
			return
		}

		h.log.Error("error authorize", "error:", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "cannot authorize user"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Status: true})
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

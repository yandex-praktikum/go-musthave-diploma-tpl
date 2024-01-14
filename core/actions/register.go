package actions

import (
	"context"
	"errors"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"net/http"

	"github.com/google/uuid"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
)

func Register(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Register started")

	body := entities.RegisterRequestBody{}
	if err := body.ParseFromRequest(req); err != nil {
		http.Error(rw, "failed parse body", http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()
	if err := db.RegisterUser(ctx, models.RegisterData{
		ID:       userID,
		Username: body.Login,
		Password: protect(body.Password),
	}); err != nil {
		if errors.Is(err, store.ErrorUniqueViolation{}) {
			http.Error(rw, "user exists", http.StatusConflict)
			return
		}
		http.Error(rw, "failed register user", http.StatusBadRequest)
		return
	}

	if err := auth.SetTokenToResponse(ctx, rw, userID); err != nil {
		http.Error(rw, "failed set token", http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

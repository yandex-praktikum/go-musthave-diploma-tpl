package actions

import (
	"context"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"net/http"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
)

func Login(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Login started")

	body := entities.LoginRequestBody{}
	if err := body.ParseFromRequest(req); err != nil {
		http.Error(rw, "failed parse body", http.StatusBadRequest)
		return
	}

	userID, err := db.LoginUser(ctx, models.LoginData{
		Username: body.Login,
		Password: protect(body.Password),
	})
	if err != nil {
		http.Error(rw, "failed login user", http.StatusBadRequest)
		return
	}

	if err = auth.SetTokenToResponse(ctx, rw, userID); err != nil {
		http.Error(rw, "failed set token", http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

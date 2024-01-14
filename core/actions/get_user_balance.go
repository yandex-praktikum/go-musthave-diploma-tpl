package actions

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
)

func GetUserBalance(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Get user balance started")

	userID := auth.FromContext(ctx)
	if userID == "" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := db.GetUserBalance(ctx, models.GetUserBalanceData{UserID: userID})
	if err != nil {
		lg.Err(err).Msg("error from store")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	result := entities.NewGetUserBalanceResponseBodyData(balance)

	b, err := json.Marshal(result)
	if err != nil {
		lg.Err(err).Msg("error marshal")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	_, err = rw.Write(b)
	if err != nil {
		lg.Err(err).Msg("error write")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

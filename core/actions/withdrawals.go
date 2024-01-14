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

func Withdrawals(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Get withdrawals started")

	userID := auth.FromContext(ctx)
	if userID == "" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdraws, err := db.Withdrawals(ctx, models.WithdrawalsData{
		UserID: userID,
	})

	if err != nil {
		lg.Err(err).Msg("error from store")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdraws.Data) == 0 {
		lg.Err(err).Msg("no withdraws")
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	result := entities.NewGetWithdrawsResponseBody(withdraws)

	//sort.Slice(result.Data, func(i, j int) bool {
	//	// although compare string is not a good idea,
	//	// I'm going to use date-time struct later.
	//	// @TODO
	//	return result.Data[i].UploadedAt < result.Data[j].UploadedAt
	//})

	b, err := json.Marshal(result.Data)
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

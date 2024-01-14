package actions

import (
	"context"
	"errors"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
	"net/http"
)

func Withdraw(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Get user withdraw started")

	userID := auth.FromContext(ctx)
	if userID == "" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	body := entities.WithdrawRequestData{}
	if err := body.ParseFromRequest(req); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := db.Withdraw(ctx, models.WithdrawData{
		UserID:  userID,
		OrderID: body.OrderID,
		Sum:     body.Sum,
	}); err != nil {
		if errors.Is(err, store.UserNoMoney{}) {
			lg.Err(err).Msg("no money")
			rw.WriteHeader(http.StatusPaymentRequired)
			return
		}

		lg.Err(err).Msg("error from store")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

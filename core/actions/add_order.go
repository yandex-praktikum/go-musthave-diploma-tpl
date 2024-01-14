package actions

import (
	"context"
	"errors"
	"net/http"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	appLog "github.com/k-morozov/go-musthave-diploma-tpl/components/logger"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
)

func AddOrder(rw http.ResponseWriter, req *http.Request, db store.Store) {
	ctx, cancel := context.WithTimeout(req.Context(), config.DefaultContextTimeout)
	defer cancel()

	lg := appLog.FromContext(ctx).With().Caller().Logger()

	lg.Debug().Msg("Add order started")

	userID := auth.FromContext(ctx)
	if userID == "" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	body := entities.AddOrderRequestBody{}
	if err := body.ParseFromRequest(req); err != nil {
		lg.Err(err).Msg("failed parse body")
		rw.WriteHeader(http.StatusBadRequest)

		return
	}

	err := db.AddOrder(ctx, models.NewAddOrderData(userID, body.OrderID))
	if err == nil {
		rw.WriteHeader(http.StatusAccepted)
		return
	}

	if !errors.Is(err, store.ErrorUniqueViolation{}) {
		lg.Err(err).Msg("error from store")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	ownerUserID, err := db.GetOwnerForOrder(ctx, models.GetOwnerForOrderData{
		OrderID: body.OrderID,
	})
	if err != nil {
		lg.Err(err).Msg("error from store")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch ownerUserID {
	case userID:
		lg.Info().
			Str("user_id", ownerUserID).
			Str("order_id", body.OrderID).
			Msg("current user created this order")
		rw.WriteHeader(http.StatusOK)
	default:
		lg.Info().
			Str("user_id", ownerUserID).
			Str("order_id", body.OrderID).
			Msg("another user is owner this order")
		rw.WriteHeader(http.StatusConflict)
	}
}

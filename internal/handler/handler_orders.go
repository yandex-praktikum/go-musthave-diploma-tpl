package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	apperror "github.com/AnxVit/go-musthave-diploma-tpl/internal/errors"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
)

// Хэндлер создания заказа пользователя
func (h *Handler) handlePostOrders(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	var req RequestOrder

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return apperror.NewBadRequestError("invalid request", "couldn't parse json")
	}

	ctx := r.Context()

	exists, err := h.service.ConfirmOrder(ctx, req.Order)
	if err != nil {
		if errors.Is(err, model.ErrorWrongOrders) {
			return apperror.NewUnprocessableEntityError(err.Error())
		}
		if errors.Is(err, model.ErrorOrderConflict) {
			return apperror.NewConflictError(err.Error())
		}
		return apperror.NewInternalServerError(err.Error())
	}

	status := http.StatusAccepted
	if exists {
		status = http.StatusOK
	}

	SendJSONResponse(w, "", status)
	return nil
}

// Хэндлер получения заказов пользователя
func (h *Handler) handleGetOrders(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	ctx := r.Context()

	orders, err := h.service.GetOrdersOfUser(ctx)
	if err != nil {
		return apperror.NewInternalServerError(err.Error())
	}

	if len(orders) == 0 {
		SendJSONResponse(w, nil, http.StatusNoContent)
		return nil
	}

	SendJSONResponse(w, orders, http.StatusOK)

	return nil
}

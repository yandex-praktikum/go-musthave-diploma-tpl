package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	apperror "github.com/AnxVit/go-musthave-diploma-tpl/internal/errors"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
)

// Хэндлер получение баланса пользователя
func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	ctx := r.Context()

	balance, err := h.service.GetBalance(ctx)
	if err != nil {
		return apperror.NewInternalServerError(err.Error())
	}

	SendJSONResponse(w, balance, http.StatusOK)

	return nil
}

// Хэндлер запроса на списание средств
func (h *Handler) handleBalanceWithdraw(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	var req BalanceWithdrawRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return apperror.NewBadRequestError("invalid request", "couldn't parse json")
	}

	ctx := r.Context()

	err = h.service.Withdrawn(ctx, &model.BalanceWithdraw{
		Order: req.Order,
		Sum:   req.Sum,
	})
	if err != nil {
		if errors.Is(err, model.ErrorOrderAlreadyExists) {
			return apperror.NewConflictError("withdrawn order already exists")
		}
		if errors.Is(err, model.ErrorInsufficientBalance) {
			return apperror.NewPaymentRequiredError(err.Error())
		}
		return apperror.NewInternalServerError(err.Error())
	}

	SendJSONResponse(w, "", http.StatusOK)

	return nil
}

// Хэндлер получения информации о выводе средств
func (h *Handler) handleWithdraw(w http.ResponseWriter, r *http.Request) *apperror.AppError {
	ctx := r.Context()

	withdrawns, err := h.service.GetWithdrawns(ctx)
	if err != nil {
		return apperror.NewInternalServerError(err.Error())
	}

	if len(withdrawns) == 0 {
		SendJSONResponse(w, "", http.StatusNoContent)
		return nil
	}

	SendJSONResponse(w, withdrawns, http.StatusOK)
	return nil
}

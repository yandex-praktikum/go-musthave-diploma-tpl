package handler

import (
	"net/http"
	"strconv"

	"github.com/brisk84/gofemart/domain"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	order, err := parseRequest[domain.Order](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusOK)
		return
	}

	order, err = h.useCase.CreateOrder(r.Context(), order)
	if err != nil {
		h.logger.Error("CreateOrder failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusOK)
		return
	}

	sendResponse(w, order, nil, http.StatusOK)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderID, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		h.logger.Error("parse order id", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusOK)
		return
	}

	user, err := h.useCase.GetOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Error("GetOrder failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusOK)
		return
	}

	sendResponse(w, user, nil, http.StatusOK)
}

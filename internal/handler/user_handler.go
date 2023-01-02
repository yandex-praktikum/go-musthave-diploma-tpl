package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/brisk84/gofemart/domain"
	"github.com/brisk84/gofemart/pkg/luhn"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest[domain.User](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if !user.IsValid() {
		sendResponse[NilType](w, nil, nil, http.StatusBadRequest)
		return
	}

	token, err := h.useCase.Register(r.Context(), user)
	if err != nil {
		if errors.Is(err, domain.ErrLoginIsBusy) {
			sendResponse[NilType](w, nil, nil, http.StatusConflict)
			return
		}
		h.logger.Error("Register failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", "Bearer "+token)

	sendResponse[NilType](w, nil, nil, http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest[domain.User](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if !user.IsValid() {
		sendResponse[NilType](w, nil, nil, http.StatusBadRequest)
		return
	}

	success, token, err := h.useCase.Login(r.Context(), user)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	if !success {
		sendResponse[NilType](w, nil, err, http.StatusUnauthorized)
		return
	}

	w.Header().Add("Authorization", "Bearer "+token)
	sendResponse[NilType](w, nil, nil, http.StatusOK)
}

func (h *Handler) UserOrdersPost(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

	user, err := h.useCase.Auth(r.Context(), token)
	if err != nil {
		h.logger.Error("Auth failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if user == nil {
		sendResponse[NilType](w, nil, nil, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("read request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	order, err := strconv.Atoi(string(body))
	if err != nil {
		sendResponse[NilType](w, nil, nil, http.StatusBadRequest)
		return
	}
	if !luhn.Valid(order) {
		sendResponse[NilType](w, nil, nil, http.StatusUnprocessableEntity)
		return
	}

	err = h.useCase.UserOrders(r.Context(), *user, order)
	if err != nil {
		if errors.Is(err, domain.ErrLoadedByThisUser) {
			sendResponse[NilType](w, nil, nil, http.StatusOK)
			return
		}
		if errors.Is(err, domain.ErrLoadedByAnotherUser) {
			sendResponse[NilType](w, nil, nil, http.StatusConflict)
			return
		}

		h.logger.Error("UserOrders failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	sendResponse[NilType](w, nil, err, http.StatusAccepted)
}

func (h *Handler) UserOrdersGet(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

	user, err := h.useCase.Auth(r.Context(), token)
	if err != nil {
		h.logger.Error("Auth failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if user == nil {
		sendResponse[NilType](w, nil, nil, http.StatusUnauthorized)
		return
	}

	orders, err := h.useCase.UserOrdersGet(r.Context(), *user)
	if err != nil {
		h.logger.Error("UserOrdersGet failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	if len(orders) < 1 {
		sendResponse[NilType](w, nil, nil, http.StatusNoContent)
		return
	}

	sendResponse[NilType](w, orders, nil, http.StatusOK)
}

func (h *Handler) UserBalanceGet(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

	user, err := h.useCase.Auth(r.Context(), token)
	if err != nil {
		h.logger.Error("Auth failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if user == nil {
		sendResponse[NilType](w, nil, nil, http.StatusUnauthorized)
		return
	}

	bal := domain.Balance{
		// Current:   decimal.NewFromFloat(500.5),
		// Withdrawn: decimal.NewFromFloat(42),
		Current:   500.5,
		Withdrawn: 42,
	}

	sendResponse[NilType](w, bal, nil, http.StatusOK)
}

func (h *Handler) UserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

	user, err := h.useCase.Auth(r.Context(), token)
	if err != nil {
		h.logger.Error("Auth failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if user == nil {
		sendResponse[NilType](w, nil, nil, http.StatusUnauthorized)
		return
	}

	withdraw, err := parseRequest[domain.Withdraw](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	order, err := strconv.Atoi(withdraw.Order)
	if err != nil || !luhn.Valid(order) {
		sendResponse[NilType](w, nil, nil, http.StatusUnprocessableEntity)
		return
	}

	err = h.useCase.UserBalanceWithdraw(r.Context(), *user, withdraw)
	if err != nil {
		h.logger.Error("UserBalanceWithdraw failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	sendResponse[NilType](w, nil, nil, http.StatusOK)
}

func (h *Handler) UserWithdrawals(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

	user, err := h.useCase.Auth(r.Context(), token)
	if err != nil {
		h.logger.Error("Auth failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}
	if user == nil {
		sendResponse[NilType](w, nil, nil, http.StatusUnauthorized)
		return
	}

	withdrawals := []domain.Withdraw{
		{
			Order: "123",
			Sum:   321,
		},
	}

	if len(withdrawals) < 1 {
		sendResponse[NilType](w, nil, nil, http.StatusNoContent)
		return
	}

	sendResponse[NilType](w, withdrawals, nil, http.StatusOK)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest[domain.User](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	user, err = h.useCase.CreateUser(r.Context(), user)
	if err != nil {
		h.logger.Error("CreateUser failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	sendResponse(w, user, nil, http.StatusOK)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		h.logger.Error("parse user id", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	user, err := h.useCase.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("GetUser failed", zap.Error(err))
		sendResponse[NilType](w, nil, err, http.StatusInternalServerError)
		return
	}

	sendResponse(w, user, nil, http.StatusOK)
}

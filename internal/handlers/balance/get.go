package balance

import (
	"context"
	"encoding/json"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"net/http"
)

type Handler struct {
	ctx     context.Context
	storage *orders.Service
	log     *logger.Logger
}

func NewHandler(ctx context.Context, storage *orders.Service, log *logger.Logger) *Handler {
	return &Handler{
		ctx:     ctx,
		storage: storage,
		log:     log,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Info("Error: not userID")
		return
	}

	balance, err := h.storage.GetBalanceUser(login)

	if err != nil {
		h.log.Error("error get balance", "error: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "error get balance"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	h.log.Info("Information req balance", "current", balance.Current, "withdraw", balance.Withdraw)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(balance); err != nil {
		h.log.Error("error get balance", "failed to marshal response: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "failed to marshal response"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)

		return
	}
}

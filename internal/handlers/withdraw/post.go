package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/utils"
	"net/http"
	"strings"
	"time"
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

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	body := RequestBody{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.log.Error("error withdraw", "error", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	// Проверяем корректность номера заказа с помощью алгоритма Луна
	orderNumber := strings.TrimSpace(body.Order)
	if !utils.IsLunaValid(orderNumber) {
		h.log.Error("error post withdraw", "error", "invalid order numbers")
		apiError, _ := json.Marshal(customerrors.APIError{Message: "invalid order numbers"})
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(apiError)
		return
	}

	// берем login пользователя
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Info("Error: not userID")
		return
	}

	now := time.Now()

	// делаем запрос в базу
	if err := h.storage.CheckWriteOffOfFunds(h.ctx, login, body.Order, body.Sum, now); err != nil {
		if errors.Is(err, customerrors.ErrNotData) {
			h.log.Error("error withdraw", "error: ", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect order"})
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write(apiError)
			return
		}

		if errors.Is(err, customerrors.ErrNotEnoughBonuses) {
			h.log.Error("error withdraw", "error: ", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "you don't have enough bonuses"})
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(apiError)
			return
		}

		h.log.Error("error withdraw", "error: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: ""})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	apiResponse, _ := json.Marshal(ResponseBody{Processing: true})
	w.WriteHeader(http.StatusOK)
	w.Write(apiResponse)

}

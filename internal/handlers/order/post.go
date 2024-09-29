package order

import (
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	service *orders.Service
	log     *logger.Logger
}

func NewHandler(service *orders.Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	// Считываем тело запроса и записываем в body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("error post order", "error:", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	// Проверяем корректность номера заказа с помощью алгоритма Луна
	orderNumber := strings.TrimSpace(string(body))
	if !utils.IsLunaValid(orderNumber) {
		h.log.Error("error post order", "error:", "invalid order numbers")
		apiError, _ := json.Marshal(customerrors.APIError{Message: "invalid order numbers"})
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(apiError)
		return
	}

	//получаем из Context token пользователя
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Error("Error post order = not userID")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//время создания запроса
	now := time.Now()

	// Проверяем заказ в базе
	if err = h.service.GetUserByAccessToken(orderNumber, login, now); err != nil {
		if errors.Is(err, customerrors.ErrAnotherUsersOrder) {
			h.log.Error("error post order", "error:", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "order number has already been uploaded by another user"})
			w.WriteHeader(http.StatusConflict)
			w.Write(apiError)
			return
		}

		if errors.Is(err, customerrors.ErrOrderIsAlready) {
			h.log.Error("error post order", "error:", err)
			apiError, _ := json.Marshal(customerrors.APIError{Message: "order number has already been uploaded by this user"})
			w.WriteHeader(http.StatusOK)
			w.Write(apiError)
			return
		}

		h.log.Error("error post order", "error:", err)
		apoError, _ := json.Marshal(customerrors.APIError{Message: "cannot loading order"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apoError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Processing: true, Order: orderNumber})
	w.WriteHeader(http.StatusAccepted)
	w.Write(response)

}

package order

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/custom_errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
	"io"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	ctx     context.Context
	service *service.Service
	log     *logger.Logger
}

func NewHandler(ctx context.Context, service *service.Service, log *logger.Logger) *Handler {
	return &Handler{
		ctx:     ctx,
		service: service,
		log:     log,
	}
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	// Считываем тело запроса и записываем в body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("error order", "error:", err)
		apiError, _ := json.Marshal(custom_errors.ApiError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	// Проверяем корректность номера заказа с помощью алгоритма Луна
	orderNumber := strings.TrimSpace(string(body))
	if !isLunaValid(orderNumber) {
		h.log.Error("error order", "error:", "invalid order numbers")
		apiError, _ := json.Marshal(custom_errors.ApiError{Message: "invalid order numbers"})
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(apiError)
		return
	}

	//получаем из Context token пользователя
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Info("Error = not userID")
	}

	//время создания запроса
	now := time.Now()

	// Проверяем заказ в базе
	err = h.service.GetUserByAccessToken(orderNumber, login, now)
	if err != nil {
		if errors.Is(err, custom_errors.ErrAnotherUsersOrder) {
			h.log.Error("error order", "error:", err)
			apiError, _ := json.Marshal(custom_errors.ApiError{Message: "order number has already been uploaded by another user"})
			w.WriteHeader(http.StatusConflict)
			w.Write(apiError)
			return
		}

		if errors.Is(err, custom_errors.ErrOrderIsAlready) {
			h.log.Error("error order", "error:", err)
			apiError, _ := json.Marshal(custom_errors.ApiError{Message: "order number has already been uploaded by this user"})
			w.WriteHeader(http.StatusOK)
			w.Write(apiError)
			return
		}

		h.log.Error("error order", "error:", err)
		apoError, _ := json.Marshal(custom_errors.ApiError{Message: "cannot loading order"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apoError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Processing: true})
	w.WriteHeader(http.StatusAccepted)
	w.Write(response)

}

// Функция, реализующая алгоритм Луна для проверки корректности номера заказа
func isLunaValid(s string) bool {
	if len(s) < 1 {
		return false
	}

	var sum, parity = 0, len(s) & 1

	for i := len(s) - 1; i >= 0; i-- {
		var digit = int(s[i] - '0')

		if parity == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		parity ^= 1
	}

	return sum%10 == 0
}

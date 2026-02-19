package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/utils"
	"go.uber.org/zap"
)

// registerOrder godoc
// @Summary Загрузка номера заказа
// @Description Сохраняет номер заказа пользователя
// @Tags orders
// @Security ApiKeyAuth
// @Accept plain
// @Produce json
// @Param order body string true "Номер заказа"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 422 {string} string "Неверный номер заказа"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/user/orders [post]
func (s *Server) registerOrder(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to get order number", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber := string(b)
	if !utils.ValidLuhn(orderNumber) {
		logger.Error("Order number is not valid")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = s.gophermart.InsertOrder(r.Context(), orderNumber)
	if err != nil {
		logger.Error("Failed to insert order", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// getOrders godoc
// @Summary Список заказов пользователя
// @Description Возвращает заказы текущего пользователя
// @Tags orders
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.OrderInfo
// @Failure 204 {string} string "Нет данных"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/user/orders [get]
func (s *Server) getOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.gophermart.GetUserOrders(r.Context())
	if err != nil {
		logger.Error("Failed to get orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	b, _ := json.Marshal(orders)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

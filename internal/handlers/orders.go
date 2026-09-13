package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/middleware"
	"github.com/b602op/go-musthave-diploma-tpl/internal/service"
)

// OrderHandler — HTTP-обработчики заказов пользователя.
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler создаёт новый OrderHandler с указанным сервисом заказов.
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// Upload обрабатывает загрузку номера заказа
func (h *OrderHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	err = h.orderService.UploadOrder(userID, orderNumber)
	if err != nil {
		switch err {
		case domain.ErrInvalidOrderNumber:
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		case domain.ErrOrderAlreadyUploadedByUser:
			w.WriteHeader(http.StatusOK)
		case domain.ErrOrderAlreadyUploadedByAnotherUser:
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// List возвращает список заказов пользователя
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Конвертируем в response DTO
	response := make([]domain.OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = domain.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"gophermart/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateOrder(w, r)
	case http.MethodGet:
		h.ListOrders(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	result, err := h.orderService.CreateOrder(r.Context(), userID, number)
	if err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case service.ErrInvalidOrderNumber:
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case service.ErrConflict:
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if result.Created {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	orders, err := h.orderService.ListOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type orderResponse struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float64 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

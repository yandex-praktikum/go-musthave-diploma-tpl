package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/eac0de/gophermart/internal/errors"
	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/middlewares"
)

type OrderHandlers struct {
	orderService *services.OrderService
}

func NewOrderHandlers(orderService *services.OrderService) *OrderHandlers {
	return &OrderHandlers{
		orderService: orderService,
	}
}

func (oh *OrderHandlers) OrderHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := middlewares.GetUserFromRequest(r)
	if r.Method == http.MethodGet {
		orders, err := oh.orderService.GetUserOrders(r.Context(), user.ID)
		if err != nil {
			message, statusCode := custom_errors.GetMessageAndStatusCode(err)
			http.Error(w, message, statusCode)
			return
		}
		if len(orders) == 0 {
			http.Error(w, "orders not found", http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orders)
		return
	} else if r.Method == http.MethodPost {
		orderNumber, err := io.ReadAll(r.Body)
		if err != nil {
			message, statusCode := custom_errors.GetMessageAndStatusCode(err)
			http.Error(w, message, statusCode)
			return
		}
		order, err := oh.orderService.AddOrder(r.Context(), string(orderNumber), user.ID)
		if err != nil {
			message, statusCode := custom_errors.GetMessageAndStatusCode(err)
			http.Error(w, message, statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(order)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

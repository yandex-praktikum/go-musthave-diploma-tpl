package order

import (
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/orders"

	"io"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	order orders.Order
}

func NewHandler(order orders.Order) *Handler {
	return &Handler{
		order: order,
	}
}

func (h Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	h.order.GetOrders()
	w.WriteHeader(http.StatusOK)
}

func (h Handler) postOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(string)
	if !ok {
		http.Error(w, "invalid user", http.StatusBadRequest)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
	}

	status := h.order.CreateOrder(userID, body)
	if status <= 0 {
		http.Error(w, "internal server error", models.InternalServerError)

	}

	w.WriteHeader(status)

}

func (h Handler) Order(r chi.Router) {
	r.Post("/orders", h.postOrder)
	r.Get("/orders", h.getOrder)
}

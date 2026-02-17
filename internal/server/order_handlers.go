package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) registerOrder(w http.ResponseWriter, r *http.Request) {
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

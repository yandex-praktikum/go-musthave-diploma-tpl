package server

import (
	"io"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/logger"
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
	if !validLuhn(orderNumber) {
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

func validLuhn(number string) bool {
	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		d := number[i] - '0'
		if d > 9 {
			return false
		}

		n := int(d)

		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}

		sum += n
		double = !double
	}

	return sum%10 == 0
}

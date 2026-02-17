package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/gophermart"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) registerWithdrawl(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req dto.WithdrawRequest
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !utils.ValidLuhn(req.Order) {
		logger.Error("Order number is not valid")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = s.gophermart.ProcessWithdraw(r.Context(), req)
	if errors.Is(err, gophermart.ErrNotEnoughBonuses) {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
}

func (s *Server) getWithdrawls(w http.ResponseWriter, r *http.Request) {
	withdrawls, err := s.gophermart.GetWithdraws(r.Context())
	if err != nil {
		logger.Error("Failed to get withdrawls", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	b, _ := json.Marshal(withdrawls)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := s.gophermart.GetUserBalance(r.Context())
	if err != nil {
		logger.Error("Failed go get user balance", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, _ := json.Marshal(balance)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

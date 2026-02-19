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

// registerWithdrawl godoc
// @Summary Списание бонусов
// @Description Регистрирует списание бонусов
// @Tags balance
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body dto.WithdrawRequest true "Данные для списания"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Некорректный JSON"
// @Failure 402 {string} string "Недостаточно бонусов"
// @Failure 422 {string} string "Неверный номер заказа"
// @Router /api/user/withdraw [post]
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

// getWithdrawls godoc
// @Summary История списаний
// @Description Возвращает историю списаний пользователя
// @Tags balance
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.WithdrawInfo
// @Failure 204 {string} string "Нет данных"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/user/withdrawals [get]
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

// getBalance godoc
// @Summary Баланс пользователя
// @Description Возвращает текущий баланс бонусов
// @Tags balance
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} dto.BalanceInfo
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/user/balance [get]
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

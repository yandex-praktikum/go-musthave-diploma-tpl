package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/NailUsmanov/internal/middleware"
	"github.com/NailUsmanov/internal/models"
	"github.com/NailUsmanov/internal/storage"
	"go.uber.org/zap"
)

func UserBalance(s storage.Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sugar.Infof(">>> UserBalance endpoint called")

		// Достаем номер пользователя из контекста через куки аутентификации
		userID, ok := r.Context().Value(middleware.UserLoginKey).(int)
		sugar.Infof("DEBUG: userID from context = %d (ok=%v)", userID, ok)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Используем метод GetUserBalance чтобы получить сумму баллов
		current, withdrawn, err := s.GetUserBalance(r.Context(), userID)
		if err != nil {
			sugar.Errorf("Failed check user balance: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Создаем структуру BalanceResponse и в нее сканируем наши значения для ответа
		balance := models.BalanceResponse{
			Current:   current,
			Withdrawn: withdrawn,
		}
		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(balance); err != nil {
			sugar.Errorf("error encoding response: %v", err)
			http.Error(w, "error with encoding response", http.StatusInternalServerError)
			return
		}
	})
}

func WithDraw(s storage.Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sugar.Infof(">>> WithDraw endpoint called")
		// Проверяем авторизацию пользователя
		userID, ok := r.Context().Value(middleware.UserLoginKey).(int)
		sugar.Infof("DEBUG: userID from context = %d (ok=%v)", userID, ok)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Проверка Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		// Декодируем JSON с номером заказа списания и суммой баллов в структуру WithDrawRequest
		var withDraw models.WithDrawRequest
		if err := json.NewDecoder(r.Body).Decode(&withDraw); err != nil {
			sugar.Error("cannot decode request JSON body:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Проверяем валидность номера заказа по Луну
		sugar.Infof("raw body for Luhn: %q", withDraw.NumberOrder)
		IsValid := IsValidLuhn(withDraw.NumberOrder)
		sugar.Infof("passed Luhn: %v", IsValid)
		if !IsValid {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}

		// Вызываем метод AddWithdrawOrder и добавляем заказ на списание в таблицу orders
		err := s.AddWithdrawOrder(r.Context(), userID, withDraw.NumberOrder, withDraw.Sum)
		// Если все err == nil возвращаем статус успешной обработки заказа
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Отлавливаем возможные ошибки
		switch {
		case errors.Is(err, storage.ErrOrderAlreadyUsed):
			sugar.Infof("Order number already used: %v", storage.ErrOrderAlreadyUsed)
			http.Error(w, "Order number already used", http.StatusConflict)
		case errors.Is(err, storage.ErrNotEnoughFunds):
			sugar.Infof("Not enough funds: %v", storage.ErrNotEnoughFunds)
			http.Error(w, "Not enough funds", http.StatusPaymentRequired)
			return
		default:
			sugar.Infof("AddWithdrawOrder failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}

func AllUserWithDrawals(s storage.Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sugar.Infof("AllUserWithDrawals endpoint called")

		// Извлекаем Юзера из контекста через куки
		userID, ok := r.Context().Value(middleware.UserLoginKey).(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Получаю все данные по списаниям конкретного пользователя через метод GetAllUserWithdrawals
		withdrawals, err := s.GetAllUserWithdrawals(r.Context(), userID)
		if err != nil {
			sugar.Errorf("GetOrdersByUserID failed: %v", err)
			http.Error(w, "Method GetAllUserWithdrawals has err", http.StatusInternalServerError)
			return
		}
		// Если нет заказов возвращаем 204 No Content
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Если все ок, то возвращаем JSON со списком списаний
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(withdrawals); err != nil {
			sugar.Error("error encoding response")
			http.Error(w, "error with encoding reponse", http.StatusInternalServerError)
			return
		}
	})
}

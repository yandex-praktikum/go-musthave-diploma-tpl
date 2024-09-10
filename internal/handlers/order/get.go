package order

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"net/http"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value(middleware.AccessTokenKey).(string)

	if !ok || login == "" {
		h.log.Info("Error: not userID")
	}

	// получаем список загруженных номеров заказов
	req, err := h.service.GetAllUserOrders(login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.log.Error("error order", "error: ", "no data to answer")
			apiError, _ := json.Marshal(customerrors.APIError{Message: "no data to answer"})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			w.Write(apiError)
		}

		h.log.Error("error order", "error: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "cannot loading order"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Processing: true})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	if err = json.NewEncoder(w).Encode(req); err != nil {
		h.log.Error("error order", "failed to marshal response: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "failed to marshal response"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)

		return
	}
}

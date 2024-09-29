package withdraw

import (
	"encoding/json"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"net/http"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Error("Error post order = not userID")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req, err := h.storage.GetWithdrawals(login)

	if err != nil {
		h.log.Error("error withdraw", "error: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: ""})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	if len(req) == 0 {
		h.log.Info("Information get withdrawals", "error: ", "there is not a single write-off")
		apiError, _ := json.Marshal(customerrors.APIError{Message: ""})
		w.WriteHeader(http.StatusNoContent)
		w.Write(apiError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(req); err != nil {
		h.log.Error("error balance", "failed to marshal response: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: "failed to marshal response"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

}

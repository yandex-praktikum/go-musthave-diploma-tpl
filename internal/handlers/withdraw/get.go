package withdraw

import (
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"net/http"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value(middleware.LoginContentKey).(string)
	if !ok || login == "" {
		h.log.Info("Error: not userID")
		return
	}

	req, err := h.storage.GetWithdrawals(h.ctx, login)

	if err != nil {
		if errors.Is(err, customerrors.ErrNotData) {
			h.log.Info("error withdraw", "error: ", "not content")
			apiError, _ := json.Marshal(customerrors.APIError{Message: "there are no write-offs"})
			w.WriteHeader(http.StatusNoContent)
			w.Write(apiError)
			return
		}
		h.log.Error("error withdraw", "error: ", err)
		apiError, _ := json.Marshal(customerrors.APIError{Message: ""})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	if len(req) == 0 {
		h.log.Error("error withdraw", "error: ", "there is not a single write-off")
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

package handlers

import (
	"encoding/json"
	"github.com/botaevg/gophermart/internal/apperror"
	"github.com/botaevg/gophermart/internal/config"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/service"
	"io"
	"net/http"
)

type handler struct {
	config         config.Config
	auth           service.Auth
	gophermart     service.Gophermart
	asyncExecution chan string
}

func NewHandler(config config.Config, auth service.Auth, gophermart service.Gophermart, asyncExecution chan string) *handler {
	return &handler{
		config:         config,
		auth:           auth,
		gophermart:     gophermart,
		asyncExecution: asyncExecution,
	}
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var UserAPI models.UserAPI
	if err := json.Unmarshal(b, &UserAPI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := h.auth.RegisterUser(UserAPI, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Bearer",
		Value: token,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("JWT " + token))
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var UserAPI models.UserAPI
	if err := json.Unmarshal(b, &UserAPI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.auth.AuthUser(UserAPI, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Bearer",
		Value: token,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("JWT " + token))
}

func (h *handler) LoadOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(apperror.UserID("username")).(uint)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	/*_, err := strconv.Atoi(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	number := uint(n)*/

	OrderUserID, err := h.gophermart.CheckOrder(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if OrderUserID == userID {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("already load"))
	} else if OrderUserID != 0 {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("already load another user"))
	}

	err = h.gophermart.AddOrder(string(b), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.asyncExecution <- string(b)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("accept new order"))
}

func (h *handler) GetListOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(apperror.UserID("username")).(uint)

	ListOrdersAPI, err := h.gophermart.GetListOrders(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ListOrdersAPI) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("no content"))
	}
	b, err := json.Marshal(&ListOrdersAPI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func (h *handler) BalanceUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(apperror.UserID("username")).(uint)

	balance, err := h.gophermart.BalanceUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	b, err := json.Marshal(&balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *handler) WithdrawRequest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(apperror.UserID("username")).(uint)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var withdrawnreq models.Withdraw

	if err := json.Unmarshal(b, &withdrawnreq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var accept bool
	accept, err = h.gophermart.WithdrawRequest(withdrawnreq, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !accept {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("withdrawn accept"))

}

func (h *handler) ListWithdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(apperror.UserID("username")).(uint)

	list, err := h.gophermart.ListWithdraw(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("no withdraw"))
	}
	b, err := json.Marshal(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

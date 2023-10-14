package balance

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/balance"
)

type Handler struct {
	wallet balance.Balance
}

func NewHandler(balance balance.Balance) *Handler {
	return &Handler{
		wallet: balance,
	}
}

func (h Handler) balance(w http.ResponseWriter, r *http.Request) {
	h.wallet.UserBalance()
	w.WriteHeader(http.StatusOK)
}

func (h Handler) balanceWithdraw(w http.ResponseWriter, r *http.Request) {
	h.wallet.BalanceWithdraw()
	w.WriteHeader(http.StatusOK)
}

func (h Handler) withdrawals(w http.ResponseWriter, r *http.Request) {
	h.wallet.UserWithdrawals()
	w.WriteHeader(http.StatusOK)
}

func (h Handler) Balance(r chi.Router) {
	r.Get("/balance", h.balance)
	r.Post("/balance/withdraw", h.balanceWithdraw)
	r.Post("/withdrawals", h.withdrawals)
}

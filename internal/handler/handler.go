// Package handler contains HTTP handlers for the Gophermart API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/true18/go-musthave-diploma-tpl/internal/auth"
	"github.com/true18/go-musthave-diploma-tpl/internal/money"
	"github.com/true18/go-musthave-diploma-tpl/internal/repository"
	"github.com/true18/go-musthave-diploma-tpl/internal/service"
)

const maxBodySize = 1 << 20

// Service is the business logic contract used by Handler.
type Service interface {
	Register(ctx context.Context, login string, password string) (int64, error)
	Login(ctx context.Context, login string, password string) (int64, error)
	UploadOrder(ctx context.Context, userID int64, number string) error
	ListOrders(ctx context.Context, userID int64) ([]repository.Order, error)
	Balance(ctx context.Context, userID int64) (repository.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount money.Amount) error
	ListWithdrawals(ctx context.Context, userID int64) ([]repository.Withdrawal, error)
}

// Handler serves the Gophermart HTTP API.
type Handler struct {
	service Service
	auth    *auth.Manager
}

// New creates a Handler.
func New(service Service, authManager *auth.Manager) *Handler {
	return &Handler{service: service, auth: authManager}
}

// Register handles POST /api/user/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.service.Register(r.Context(), req.Login, req.Password)
	switch {
	case err == nil:
		h.auth.SetCookie(w, userID)
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrUserExists):
		w.WriteHeader(http.StatusConflict)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Login handles POST /api/user/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.service.Login(r.Context(), req.Login, req.Password)
	switch {
	case err == nil:
		h.auth.SetCookie(w, userID)
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidCredentials):
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// UploadOrder handles POST /api/user/orders.
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !isPlainText(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.UploadOrder(r.Context(), userID, number)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, service.ErrOrderExistsSameUser):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrOrderExistsAnotherUser):
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, service.ErrInvalidOrderFormat):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidOrderNumber):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Orders handles GET /api/user/orders.
func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := h.service.ListOrders(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: formatTime(order.UploadedAt),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// Balance handles GET /api/user/balance.
func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.service.Balance(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, balanceResponse{Current: balance.Current, Withdrawn: balance.Withdrawn})
}

// Withdraw handles POST /api/user/balance/withdraw.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req withdrawRequest
	if !decodeJSONWithStatus(w, r, &req, http.StatusUnprocessableEntity) {
		return
	}
	if req.Order == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInsufficientFunds):
		w.WriteHeader(http.StatusPaymentRequired)
	case errors.Is(err, service.ErrInvalidOrderNumber), errors.Is(err, service.ErrInvalidWithdrawalAmount):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Withdrawals handles GET /api/user/withdrawals.
func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.service.ListWithdrawals(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]withdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		response = append(response, withdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: formatTime(withdrawal.ProcessedAt),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

type credentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type withdrawRequest struct {
	Order string       `json:"order"`
	Sum   money.Amount `json:"sum"`
}

type orderResponse struct {
	Number     string        `json:"number"`
	Status     string        `json:"status"`
	Accrual    *money.Amount `json:"accrual,omitempty"`
	UploadedAt string        `json:"uploaded_at"`
}

type balanceResponse struct {
	Current   money.Amount `json:"current"`
	Withdrawn money.Amount `json:"withdrawn"`
}

type withdrawalResponse struct {
	Order       string       `json:"order"`
	Sum         money.Amount `json:"sum"`
	ProcessedAt string       `json:"processed_at"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return decodeJSONWithStatus(w, r, dst, http.StatusBadRequest)
}

func decodeJSONWithStatus(w http.ResponseWriter, r *http.Request, dst any, status int) bool {
	reader := http.MaxBytesReader(w, r.Body, maxBodySize)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		w.WriteHeader(status)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func isPlainText(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return true
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "text/plain"
}

func formatTime(value time.Time) string {
	return value.Format(time.RFC3339)
}

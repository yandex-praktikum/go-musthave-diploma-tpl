package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, req models.RegisterRequest) (string, error)
	Login(ctx context.Context, req models.RegisterRequest) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type OrderService interface {
	UploadOrder(ctx context.Context, orderNumber string, userID int64) error
	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
	GetOrderAccrual(ctx context.Context, orderID int64) (*float64, error)
	GetUserBalance(ctx context.Context, userID int64) (float64, float64, error)
	WithdrawBalance(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error)
}

type Handler struct {
	UserService  UserService
	OrderService OrderService
	Logger       *zap.Logger
}

func NewHandler(userService *service.UserService, orderService *service.OrderService, logger *zap.Logger) *Handler {
	return &Handler{UserService: userService, OrderService: orderService, Logger: logger}
}

func (h *Handler) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		token, err := h.UserService.Register(r.Context(), req)
		if err != nil {
			switch err {
			case service.ErrUserExists:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		token, err := h.UserService.Login(r.Context(), req)
		if err != nil {
			switch err {
			case service.ErrUserNotFound, service.ErrInvalidPassword:
				http.Error(w, err.Error(), http.StatusUnauthorized)
			default:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) UploadOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getUserFromRequest(r)
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		orderNumberBytes := make([]byte, 64)
		n, err := r.Body.Read(orderNumberBytes)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		orderNumber := string(orderNumberBytes[:n])
		err = h.OrderService.UploadOrder(r.Context(), orderNumber, user.ID)
		if err != nil {
			switch err {
			case service.ErrInvalidOrderFormat:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case service.ErrInvalidOrderNumber:
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			case service.ErrOrderAlreadyUploadedByUser:
				w.WriteHeader(http.StatusOK)
			case service.ErrOrderAlreadyUploadedByAnother:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handler) GetOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getUserFromRequest(r)
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		orders, err := h.OrderService.GetOrdersByUserID(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := make([]models.OrderResponse, 0, len(orders))
		for _, o := range orders {
			accrual, err := h.OrderService.GetOrderAccrual(r.Context(), o.ID)
			if err != nil {
				h.Logger.Error("Ошибка получения начисления для заказа", zap.Error(err))
			}
			resp = append(resp, models.OrderResponse{
				Number:     o.OrderNumber,
				Status:     o.Status,
				Accrual:    accrual,
				UploadedAt: o.CreatedAt.Format(time.RFC3339),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (h *Handler) GetUserBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getUserFromRequest(r)
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		current, withdrawn, err := h.OrderService.GetUserBalance(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		resp := struct {
			Current   float64 `json:"current"`
			Withdrawn float64 `json:"withdrawn"`
		}{
			Current:   current,
			Withdrawn: withdrawn,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (h *Handler) WithdrawBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getUserFromRequest(r)
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		var req struct {
			Order string  `json:"order"`
			Sum   float64 `json:"sum"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusUnprocessableEntity)
			return
		}
		if req.Order == "" || req.Sum <= 0 {
			http.Error(w, "неверный номер заказа или сумма", http.StatusUnprocessableEntity)
			return
		}
		err := h.OrderService.WithdrawBalance(r.Context(), user.ID, req.Order, req.Sum)
		if err != nil {
			switch err {
			case service.ErrInsufficientFunds:
				http.Error(w, "недостаточно средств", http.StatusPaymentRequired)
			case service.ErrInvalidOrderNumber:
				http.Error(w, "неверный номер заказа", http.StatusUnprocessableEntity)
			default:
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) GetUserWithdrawalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getUserFromRequest(r)
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		withdrawals, err := h.OrderService.GetUserWithdrawals(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(withdrawals)
	}
}

func (h *Handler) getUserFromRequest(r *http.Request) (*models.User, bool) {
	userIDStr, ok := GetUserIDFromContext(r.Context())
	if !ok {
		return nil, false
	}
	user, err := h.UserService.GetUserByLogin(r.Context(), userIDStr)
	if err != nil {
		return nil, false
	}
	return user, true
}

func SetupRoutersWithLogger(h *Handler, logger *zap.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(LoggingMiddleware(logger))
	r.Post("/api/user/register", h.RegisterHandler())
	r.Post("/api/user/login", h.LoginHandler())
	r.With(AuthMiddleware).Post("/api/user/orders", h.UploadOrderHandler())
	r.With(AuthMiddleware).Get("/api/user/orders", h.GetOrdersHandler())
	r.With(AuthMiddleware).Get("/api/user/balance", h.GetUserBalanceHandler())
	r.With(AuthMiddleware).Post("/api/user/balance/withdraw", h.WithdrawBalanceHandler())
	r.With(AuthMiddleware).Get("/api/user/withdrawals", h.GetUserWithdrawalsHandler())
	return r
}

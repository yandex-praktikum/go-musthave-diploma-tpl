package routers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/service"
	"github.com/go-chi/chi/v5"
)

type UserService interface {
	Register(req models.RegisterRequest) (string, error)
	Login(req models.RegisterRequest) (string, error)
	GetUserByLogin(login string) (*models.User, error)
}

type OrderService interface {
	UploadOrder(orderNumber string, userID int64) error
}

type Handler struct {
	UserService  UserService
	OrderService OrderService
}

func NewHandler(userService *service.UserService, orderService *service.OrderService) *Handler {
	return &Handler{UserService: userService, OrderService: orderService}
}

func (h *Handler) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		token, err := h.UserService.Register(req)
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
		token, err := h.UserService.Login(req)
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
		userIDStr, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}
		user, err := h.UserService.GetUserByLogin(userIDStr)
		if err != nil {
			http.Error(w, "пользователь не найден", http.StatusUnauthorized)
			return
		}
		orderNumberBytes := make([]byte, 64)
		n, err := r.Body.Read(orderNumberBytes)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, "неверный формат запроса", http.StatusBadRequest)
			return
		}
		orderNumber := string(orderNumberBytes[:n])
		err = h.OrderService.UploadOrder(orderNumber, user.ID)
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

func SetupRouters(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/user/register", h.RegisterHandler())
	r.Post("/api/user/login", h.LoginHandler())
	r.With(AuthMiddleware).Post("/api/user/orders", h.UploadOrderHandler())
	return r
}

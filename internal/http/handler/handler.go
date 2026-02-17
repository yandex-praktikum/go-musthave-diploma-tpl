// Package handler реализует HTTP хендлеры
package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/anon-d/gophermarket/internal/http/middleware"
	"github.com/anon-d/gophermarket/internal/http/service"
)

// AuthRequest структура запроса на регистрацию и логин
type AuthRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// WithdrawRequest структура запроса на списание баллов
type WithdrawRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required"`
}

// OrderResponse структура ответа с заказом
type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

// BalanceResponse структура ответа с балансом
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalResponse структура ответа со списанием
type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// GopherHandler HTTP хендлер
type GopherHandler struct {
	service *service.GopherService
}

func NewGopherHandler(svc *service.GopherService) *GopherHandler {
	return &GopherHandler{service: svc}
}

// RegisterUser хендлер запроса на регистрацию и, в случае успеха, аутентификации пользователя
func (h *GopherHandler) RegisterUser(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := h.service.RegisterUser(c.Request.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.Status(http.StatusConflict)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	h.setAuthCookie(c, token)
	c.Status(http.StatusOK)
}

// LoginUser хендлер запроса на аутентификацию пользователя
func (h *GopherHandler) LoginUser(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := h.service.LoginUser(c.Request.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.Status(http.StatusUnauthorized)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	h.setAuthCookie(c, token)
	c.Status(http.StatusOK)
}

// CreateOrder хендлер запроса на загрузку пользователем номера заказа для расчёта
func (h *GopherHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	err = h.service.CreateOrder(c.Request.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, service.ErrOrderExists) {
			c.Status(http.StatusOK) // заказ уже был загружен этим пользователем
			return
		}
		if errors.Is(err, service.ErrOrderExistsByAnotherUser) {
			c.Status(http.StatusConflict) // заказ уже был загружен другим пользователем
			return
		}
		if errors.Is(err, service.ErrInvalidOrderNumber) {
			c.Status(http.StatusUnprocessableEntity) // неверный формат номера заказа
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusAccepted) // новый номер заказа принят в обработку
}

// GetOrders хендлер запроса на получение списка загруженных пользователем номеров заказов,
// статусов их обработки и информации о начислениях
func (h *GopherHandler) GetOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrders(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, order := range orders {
		resp := OrderResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		if order.Accrual > 0 {
			accrual := order.Accrual
			resp.Accrual = &accrual
		}
		response[i] = resp
	}

	c.JSON(http.StatusOK, response)
}

// GetBalance хендлер запроса на получение текущего баланса счёта баллов лояльности пользователя
func (h *GopherHandler) GetBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}

// WithdrawBalance хендлер запроса на списание баллов с накопительного счёта
// в счёт оплаты нового заказа
func (h *GopherHandler) WithdrawBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	var request WithdrawRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.service.Withdraw(c.Request.Context(), userID, request.Order, request.Sum)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientFunds) {
			c.Status(http.StatusPaymentRequired) // на счету недостаточно средств
			return
		}
		if errors.Is(err, service.ErrInvalidOrderNumber) {
			c.Status(http.StatusUnprocessableEntity) // неверный номер заказа
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// GetWithdrawals хендлер запроса на получение информации о выводе средств
// с накопительного счёта пользователем
func (h *GopherHandler) GetWithdrawals(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.service.GetWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	response := make([]WithdrawalResponse, len(withdrawals))
	for i, w := range withdrawals {
		response[i] = WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}

// setAuthCookie устанавливает куки с токеном и заголовок Authorization
// неизвестно что будет использовать пользователь, использую оба
func (h *GopherHandler) setAuthCookie(c *gin.Context, token string) {
	c.SetCookie("token", token, 86400, "/", "", false, true)
	c.Header("Authorization", "Bearer "+token)
}

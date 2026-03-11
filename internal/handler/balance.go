package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/middleware"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
)

// GetBalance — GET /api/user/balance (только для аутентифицированных).
func (h *Handler) GetBalance(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	balance, err := h.Services.Balance.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		Current:   float64(balance.Current),
		Withdrawn: float64(balance.Withdrawn),
	})
}

// PostWithdraw — POST /api/user/balance/withdraw (только для аутентифицированных).
func (h *Handler) PostWithdraw(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	sum := int64(req.Sum)
	if sum <= 0 || float64(sum) != req.Sum {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sum must be a positive integer"})
		return
	}

	err = h.Services.Balance.Withdraw(c.Request.Context(), userID, req.Order, sum)
	if err != nil {
		var insufficient *repository.ErrInsufficientFunds
		if errors.As(err, &insufficient) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
			return
		}
		var val *service.ErrValidation
		if errors.As(err, &val) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": val.Error()})
			return
		}
		var dup *repository.ErrDuplicateWithdrawalOrder
		if errors.As(err, &dup) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusOK)
}

// GetWithdrawals — GET /api/user/withdrawals (только для аутентифицированных).
func (h *Handler) GetWithdrawals(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	list, err := h.Services.Balance.ListWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if len(list) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	items := make([]WithdrawalItem, 0, len(list))
	for _, w := range list {
		items = append(items, WithdrawalItem{
			Order:       w.Order,
			Sum:         float64(w.Sum),
			ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, items)
}

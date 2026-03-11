package handler

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/middleware"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
)

// PostOrders — POST /api/user/orders (только для аутентифицированных).
func (h *Handler) PostOrders(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	order, created, err := h.Services.Order.AddOrder(c.Request.Context(), userID, string(body))
	if err != nil {
		var val *service.ErrValidation
		if errors.As(err, &val) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": val.Error()})
			return
		}
		var other *service.ErrOrderOwnedByOther
		if errors.As(err, &other) {
			c.JSON(http.StatusConflict, gin.H{"error": "order already uploaded by another user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	_ = order
	if created {
		c.Status(http.StatusAccepted)
		return
	}
	c.Status(http.StatusOK)
}

// GetOrders — GET /api/user/orders (только для аутентифицированных).
func (h *Handler) GetOrders(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	orders, err := h.Services.Order.ListOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	items := make([]OrderItem, 0, len(orders))
	for _, o := range orders {
		item := OrderItem{
			Number:     o.Number,
			Status:     o.Status,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		}
		if o.Accrual != nil {
			acc := float64(*o.Accrual)
			item.Accrual = &acc
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, items)
}

package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/auth"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/service"
)

// Register — POST /api/user/register.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.Services.User.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		var dup *repository.ErrDuplicateLogin
		if errors.As(err, &dup) {
			c.JSON(http.StatusConflict, gin.H{"error": "login already taken"})
			return
		}
		var val *service.ErrValidation
		if errors.As(err, &val) {
			c.JSON(http.StatusBadRequest, gin.H{"error": val.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	auth.SetAuthCookie(c.Writer, user.ID, h.CookieSecret)
	c.Status(http.StatusOK)
}

// Login — POST /api/user/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.Services.User.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		var invalid *service.ErrInvalidCredentials
		if errors.As(err, &invalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		var val *service.ErrValidation
		if errors.As(err, &val) {
			c.JSON(http.StatusBadRequest, gin.H{"error": val.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	auth.SetAuthCookie(c.Writer, user.ID, h.CookieSecret)
	c.Status(http.StatusOK)
}

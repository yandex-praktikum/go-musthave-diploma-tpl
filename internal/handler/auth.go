package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

func (h *Handler) SingUp(c *gin.Context) {
	var input models.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, err)
		return
	}
	input.Login = validatelogin(input.Login)

	_, err := h.authService.CreateUser(input)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	token, err := h.authService.GenerateToken(input.Login, input.Password)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Writer.Header().Set("Authorization", token)
	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})

}

func (h *Handler) SingIn(c *gin.Context) {
	var input models.User
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, err)
		return
	}
	input.Login = validatelogin(input.Login)

	token, err := h.authService.GenerateToken(input.Login, input.Password)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Writer.Header().Set("Authorization", token)

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})

}

func validatelogin(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

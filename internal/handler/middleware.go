package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	hashHeader = "Authorization"
	userCtx    = "userId"
)

func (h *Handler) UserIdentify(c *gin.Context) {

	header := c.GetHeader(hashHeader)

	if header == "" {
		newErrorResponse(c, errors.New("unauthorized"))
		return
	}

	userID, err := h.authService.ParseToken(header)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Set(userCtx, userID)
}

func getUserID(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	unauthorizedErr := errors.New("Unauthorized")
	if !ok {
		newErrorResponse(c, unauthorizedErr)
		return 0, errors.New("user id not found")
	}
	idInt, ok := id.(int)
	if !ok {
		newErrorResponse(c, unauthorizedErr)
		return 0, errors.New("user id not found")
	}

	return idInt, nil
}

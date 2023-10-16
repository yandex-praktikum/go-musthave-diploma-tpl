package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/constants"
)

func (h *Handler) UserIdentify(c *gin.Context) {

	header := c.GetHeader(constants.HashHeader)

	if header == "" {
		newErrorResponse(c, errors.New("unauthorized"))
		return
	}

	userID, err := h.service.Autorisation.ParseToken(header)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Set(constants.UserCtx, userID)
}

func getUserID(c *gin.Context) (int, error) {
	id, ok := c.Get(constants.UserCtx)
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

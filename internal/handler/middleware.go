package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/constants"
)

func (h *Handler) UserIdentify(c *gin.Context) {

	header := c.GetHeader(constants.HashHeader)

	if header == "" {
		newErrorResponse(c, errors.New("unauthorized"))
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, errors.New("unauthorized"))
		return
	}

	userId, err := h.storage.Autorisation.ParseToken(headerParts[1])
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Set(constants.UserCtx, userId)
}

func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(constants.UserCtx)
	unauthorized_err := errors.New("Unauthorized")
	if !ok {
		newErrorResponse(c, unauthorized_err)
		return 0, errors.New("user id not found")
	}
	idInt, ok := id.(int)
	if !ok {
		newErrorResponse(c, unauthorized_err)
		return 0, errors.New("user id not found")
	}

	return idInt, nil
}

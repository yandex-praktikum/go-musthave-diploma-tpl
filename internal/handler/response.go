package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/httperrors"
)

func newErrorResponse(c *gin.Context, err error) {
	er := httperrors.ParseErrors(err, true)
	c.AbortWithStatusJSON(er.Status(), errorResponse{er.Error()})
}

type errorResponse struct {
	Message string `json:"message"`
}

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/httperrors"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

func newErrorResponse(c *gin.Context, err error) {
	er := httperrors.ParseErrors(err, true)
	c.AbortWithStatusJSON(er.Status(), errorResponse{er.Error()})
}

type errorResponse struct {
	Message string `json:"message"`
}

type getAllOrdersResponse struct {
	Data []models.Order `json:"data"`
}

type getAllWithdrawalsResponse struct {
	Data []models.WithdrawResponse `json:"data"`
}

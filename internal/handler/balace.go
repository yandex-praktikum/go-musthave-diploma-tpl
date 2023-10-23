package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

func (h *Handler) GetBalance(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	currentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	balance, err := h.balanceService.GetBalance(currentuserID)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, balance)
}

func (h *Handler) Withdraw(c *gin.Context) {

	c.Writer.Header().Set("Content-Type", "application/json")
	currentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	var withdraw models.Withdraw

	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponse(c, err)
		h.log.Error(err)
		return
	}

	defer c.Request.Body.Close()

	if err := json.Unmarshal(jsonData, &withdraw); err != nil {
		newErrorResponse(c, err)
		h.log.Error(err)
		return
	}

	err = h.balanceService.Withdraw(currentuserID, withdraw)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, "bonuses was debeted")
}

func (h *Handler) GetWithdraws(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	currentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	withdraws, err := h.balanceService.GetWithdraws(currentuserID)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if len(withdraws) == 0 {
		newErrorResponse(c, errors.New("NoContent"))
		return
	}
	c.JSON(http.StatusOK, withdraws)
}

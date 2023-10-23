package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/luhn"
)

func (h *Handler) GetOrders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	currentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	orders, err := h.ordersService.GetOrders(currentuserID)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if len(orders) == 0 {
		newErrorResponse(c, errors.New("NoContent"))
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) PostOrder(c *gin.Context) {
	fmt.Println("1111111111111111")
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponse(c, err)
		h.log.Error(err)
		return
	}
	defer c.Request.Body.Close()

	numOrder := string(data)
	numOrderInt, err := strconv.Atoi(numOrder)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	correctnum := luhn.Valid(numOrderInt)

	if !correctnum {
		newErrorResponse(c, errors.New("UnprocessableEntity"))
		return
	}

	currentuserID, err := getUserID(c)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	userID, updatedate, err := h.ordersService.CreateOrder(currentuserID, numOrder, "NEW")

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if currentuserID != userID {
		newErrorResponse(c, errors.New("conflict"))
		return
	}

	if currentuserID == userID && !updatedate.IsZero() {
		c.JSON(http.StatusOK, "Order was save earlier")
		return
	}

	c.JSON(http.StatusAccepted, "Order saved")
}

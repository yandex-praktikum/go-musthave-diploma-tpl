package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/luhn"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

func (h *Handler) GetOrders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	curentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	orders, err := h.storage.GetOrders(curentuserID)
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

	curentuserID, err := getUserID(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	userID, apdatedate, err := h.storage.Orders.CreateOrder(curentuserID, numOrder, "NEW")

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if curentuserID != userID {
		newErrorResponse(c, errors.New("conflict"))
		return
	}

	if curentuserID == userID && !apdatedate.IsZero() {
		c.JSON(http.StatusOK, "Order was save earlier")
		return
	}

	c.JSON(http.StatusAccepted, "Order saved")
}

type OrdersService struct {
	// log  logger.Logger
	repo repository.Orders
}

func (r *OrdersService) GetOrdersWithStatus() ([]models.OrderResponse, error) {
	return r.repo.GetOrdersWithStatus()
}

func (r *OrdersService) ChangeStatusAndSum(sum float64, status, num string) error {
	return r.repo.ChangeStatusAndSum(sum, status, num)
}

func (r *OrdersService) CreateOrder(userID int, num, status string) (int, time.Time, error) {
	return r.repo.CreateOrder(userID, num, status)
}

func (r *OrdersService) GetOrders(userID int) ([]models.Order, error) {
	return r.repo.GetOrders(userID)
}

func NewOrdersStorage(repo repository.Orders) *OrdersService {
	return &OrdersService{repo: repo}
}

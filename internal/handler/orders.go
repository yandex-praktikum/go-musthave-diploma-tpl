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
	curentuserId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	orders, err := h.storage.GetOrders(curentuserId)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if len(orders) == 0 {
		newErrorResponse(c, errors.New("NoContent"))
		return
	}
	c.JSON(http.StatusOK, getAllOrdersResponse{
		Data: orders,
	})
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

	curentuserId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	user_id, apdatedate, err := h.storage.Orders.CreateOrder(curentuserId, numOrder, "NEW")

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	if curentuserId != user_id {
		newErrorResponse(c, errors.New("conflict"))
		return
	}

	if curentuserId == user_id && !apdatedate.IsZero() {
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

func (r *OrdersService) CreateOrder(user_id int, num, status string) (int, time.Time, error) {
	return r.repo.CreateOrder(user_id, num, status)
}

func (r *OrdersService) GetOrders(user_id int) ([]models.Order, error) {
	return r.repo.GetOrders(user_id)
}

func NewOrdersStorage(repo repository.Orders) *OrdersService {
	return &OrdersService{repo: repo}
}

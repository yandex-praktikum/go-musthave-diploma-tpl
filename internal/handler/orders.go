package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	numOrderInt, err := strconv.Atoi(c.Param("numorder"))

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	curentuserId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	user_id, apdatedate, err := h.storage.Orders.CreateOrder(numOrderInt, curentuserId, "New")

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

func (r *OrdersService) CreateOrder(num, user_id int, status string) (int, time.Time, error) {
	return r.repo.CreateOrder(num, user_id, status)
}

func (r *OrdersService) GetOrders(user_id int) ([]models.Order, error) {
	return r.repo.GetOrders(user_id)
}

func NewOrdersStorage(repo repository.Orders) *OrdersService {
	return &OrdersService{repo: repo}
}

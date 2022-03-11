package handler

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"Loyalty/pkg/luhn"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service   *service.Service
	UserLogin string
	logger    *logrus.Logger
}

func NewHandler(service *service.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

//=========================================================================
func (h *Handler) SaveOrder(c *gin.Context) {
	//read request body
	number, err := ioutil.ReadAll(c.Request.Body)
	if err != nil || string(number) == "" {
		c.String(http.StatusBadRequest, "Bad request")
		return
	}
	//validate order number
	if ok := luhn.Validate(string(number)); !ok {
		c.String(http.StatusUnprocessableEntity, "Not valid number of order")
		return
	}

	if err := h.service.SaveOrder(string(number), h.UserLogin); err != nil {
		h.logger.Error(err)
		switch err {
		case repository.ErrInt:
			c.String(http.StatusInternalServerError, err.Error())
		case repository.ErrOrdUsrConfl:
			c.String(http.StatusConflict, err.Error())
		case repository.ErrOrdOverLap:
			c.String(http.StatusOK, err.Error())
		default:
			c.String(http.StatusInternalServerError, err.Error())

		}
		return
	}

	c.String(http.StatusAccepted, "order has been accepted for processing")
}

//=========================================================================
func (h *Handler) GetOrders(c *gin.Context) {
	ordersList, err := h.service.GetOrders(h.UserLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(ordersList) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"Info": "Oredrs not found"})
		return
	}
	c.JSON(http.StatusOK, ordersList)
}

//=========================================================================
func (h *Handler) GetBalance(c *gin.Context) {
	accountState, err := h.service.GetBalance(h.UserLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	var balance models.Balance
	balance.Current = float64(accountState.Current) / 100
	balance.Withdrawn = float64(accountState.Withdrawn) / 100
	c.JSON(http.StatusOK, balance)
}

//=========================================================================
func (h *Handler) Withdraw(c *gin.Context) {
	var withdraw models.WithdrawalDTO
	if err := c.ShouldBindJSON(&withdraw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Withdraw(&withdraw, h.UserLogin); err != nil {
		h.logger.Error(err)
		switch err {
		case service.ErrInt:
			c.JSON(http.StatusInternalServerError, err.Error())
		case service.ErrNoMoney:
			c.JSON(http.StatusPaymentRequired, err.Error())
		case service.ErrNotValid:
			c.JSON(http.StatusUnprocessableEntity, err.Error())
		default:
			c.JSON(http.StatusInternalServerError, err.Error())

		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawal": "done"})
}

//=========================================================================
func (h *Handler) GetWithdrawals(c *gin.Context) {
	withdrawls, err := h.service.GetWithdrawals(h.UserLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(withdrawls) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"info": "withdrawls not found"})
		return
	}
	c.JSON(http.StatusOK, withdrawls)
}

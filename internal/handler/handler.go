package handler

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   *service.Service
	userLogin string
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

//=========================================================================
func (h *Handler) Init() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())

	//auth
	//registration
	router.POST("/api/user/register", h.SignIn)
	//login
	router.POST("/api/user/login", h.SignUp)
	//update token
	router.POST("/api/user/updatetoken", h.TokenRefreshing)

	user := router.Group("/api/user", h.AuthMiddleware)
	{
		//orders from user
		user.POST("/orders", h.saveOrder)
		//withdrawal request
		user.POST("/balance/withdraw", h.withdraw)
		//getting a list of orders
		user.GET("/orders", h.getOrders)
		//getting balance
		user.GET("/balance", h.getBalance)
		//getting information of withdrawals
		user.GET("/balance/withdrawals", h.getWithdrowals)
	}
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Not correct URL"})
	})
	return router
}

//=========================================================================
func (h *Handler) saveOrder(c *gin.Context) {
	//read request body
	number, err := ioutil.ReadAll(c.Request.Body)
	if err != nil || string(number) == "" {
		c.String(http.StatusBadRequest, "Bad request")
		return
	}
	//validate order number
	if ok := govalidator.IsNumeric(string(number)); !ok {
		c.String(http.StatusUnprocessableEntity, "Not valid order number")
		return
	}
	//запрос в систему начисления баллов
	//получение ответа
	//..................
	var order models.Order
	order.Number = string(number)
	order.Status = "status"
	order.Accrual = rand.Float64()
	if err := h.service.Repository.SaveOrder(&order, h.userLogin); err != nil {
		switch err.Error() {
		case repository.ErrInt.Error():
			c.String(http.StatusInternalServerError, err.Error())
		case repository.ErrOrdUsrConfl.Error():
			c.String(http.StatusConflict, err.Error())
		case repository.ErrOrdOverLap.Error():
			c.String(http.StatusOK, err.Error())
		default:
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.String(http.StatusAccepted, "order has been accepted for processing")
}

//=========================================================================
func (h *Handler) getOrders(c *gin.Context) {
	ordersList, err := h.service.Repository.GetOrders(h.userLogin)
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
func (h *Handler) getBalance(c *gin.Context) {
	accountState, err := h.service.Repository.GetBalance(h.userLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, accountState)
}

//=========================================================================
func (h *Handler) withdraw(c *gin.Context) {
	var withdraw models.Withdraw
	if err := c.ShouldBindJSON(&withdraw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//check orer in db
	status, err := h.service.Repository.CheckOrder(withdraw.Order, h.userLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	//chcek order status
	if status == "" {
		c.JSON(http.StatusUnprocessableEntity, repository.ErrInt)
		return
	}
	/////........... Если статус ??? доделать

	//check bonuses
	accountState, err := h.service.Repository.GetBalance(h.userLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// if not enough bonuses
	if accountState.Current < withdraw.Sum {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "not enough bonuses"})
		return
	}
	if err := h.service.Repository.Withdraw(&withdraw, h.userLogin); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"withdraw": "done"})
}

//=========================================================================
func (h *Handler) getWithdrowals(c *gin.Context) {
	withdrawls, err := h.service.Repository.GetWithdrawls(h.userLogin)
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

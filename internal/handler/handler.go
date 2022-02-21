package handler

import (
	"Loyalty/internal/models"
	"Loyalty/internal/service"
	"errors"
	"io/ioutil"
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
		router.POST("/balance/withdraw")
		//getting a list of orders
		router.GET("/orders")
		//getting balance
		router.GET("/balance")
		//getting information of withdrawals
		router.GET("/balance/withdrawals")
	}
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Not correct URL"})
	})
	return router
}
func (h *Handler) saveOrder(c *gin.Context) {
	//read request body
	number, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
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
	order.Accrual = 40.2
	if err := h.service.Repository.SaveOrder(&order, h.userLogin); err != nil {
		if errors.Is(err, "error: internal db error") {

		}
	}
	c.String(http.StatusAccepted, "order has been accepted for processing")
}

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/handler"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service"
)

func (s *Server) NewRouter(auth service.Autorisation, orders repository.Orders, balance service.Account) *gin.Engine {

	h := handler.NewHandler(auth, orders, balance, s.cfg, s.log)

	router := gin.New()

	router.POST("/api/user/register", h.SingUp)
	router.POST("/api/user/login", h.SingIn)

	router.POST("/api/user/orders", h.UserIdentify, h.PostOrder)
	router.GET("/api/user/orders", h.UserIdentify, h.GetOrders)

	router.GET("/api/user/balance", h.UserIdentify, h.GetBalance)
	router.POST("/api/user/balance/withdraw", h.UserIdentify, h.Withdraw)
	router.GET("/api/user/withdrawals", h.UserIdentify, h.GetWithdraws)

	return router
}

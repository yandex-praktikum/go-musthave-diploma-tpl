package router

import (
	"github.com/gin-gonic/gin"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/config"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/handler"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/middleware"
)

// SetupRouter создаёт Gin-роутер.
func SetupRouter(h *handler.Handler, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/user", middleware.Gzip())
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
		protected := api.Group("", middleware.RequireAuth(cfg.CookieSecret))
		{
			protected.POST("/orders", h.PostOrders)
			protected.GET("/orders", h.GetOrders)
			protected.GET("/balance", h.GetBalance)
			protected.POST("/balance/withdraw", h.PostWithdraw)
			protected.GET("/withdrawals", h.GetWithdrawals)
		}
	}
	return r
}

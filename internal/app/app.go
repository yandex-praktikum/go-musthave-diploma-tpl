package app

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/api"
)

func Run() {

	cfg := config.NewCfg()

	server := api.NewServer(cfg)

	fmt.Println(server.Config)
	fmt.Println(server.DB)

	e := echo.New()
	e.POST("/api/user/registrater", server.UserRegistrater)
	e.POST("/api/user/login", server.UserAuthentication)

	r := e.Group("/")
	{
		r.Use(api.JWTMiddleware)

		r.POST("api/user/orders", server.UploadOrder)
		r.GET("api/user/orders", server.GetOrders)
		r.GET("api/user/:number", server.GetOrderByParam)

		r.POST("api/user/balance/withdraw", server.WithdrawnDebitPoints)
		r.GET("api/user/balance/withdraw", server.WithdrawnInformationAboutDebitingFunds)

		r.GET("api/user/balance", server.GetUserBalace)

	}

	e.Logger.Fatal(e.Start(":8080"))
}

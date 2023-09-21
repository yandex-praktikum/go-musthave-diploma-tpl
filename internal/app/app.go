package app

import (
	"fmt"
	//"log"
	//"net/http"

	"github.com/labstack/echo/v4"

	//"github.com/go-chi/chi/v5"
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
	e.POST("/api/user/login", server.UserAuthorization)

	e.Logger.Fatal(e.Start(":8080"))
}

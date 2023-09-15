package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kindenko/gophermart/config"
	"github.com/kindenko/gophermart/internal/storage"
)

func Run() {
	e := echo.New()

	cfg := config.NewCfg()
	InitDB := storage.InitDB(*cfg)

	fmt.Println(InitDB)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

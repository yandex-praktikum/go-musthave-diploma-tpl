package main

import (
	"context"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	conf := config.NewConfig()
	parseFlags(conf)
	parseEnvs(conf)

	postgresqlURL := conf.DatabaseURI

	if postgresqlURL == "" {
		log.Fatal("no DATABASE_URI in .env")
		return
	}

	db, err := gorm.Open(postgres.Open(postgresqlURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	application.AppFactory(db, conf)

	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	go application.App.AccrualService.ProcessOrders(e)
	go application.App.AccrualService.ProcessFailedOrders(e)

	err = runMigrate(e, *conf)
	if err != nil {
		return
	}

	// middleware
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// decompress
	e.Use(middleware.Decompress())

	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		BeforeFunc: auth.BeforeFunc,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &auth.Claims{}
		},
		SigningKey:    []byte(auth.GetJWTSecret()),
		SigningMethod: jwt.SigningMethodHS256.Alg(),
		TokenLookup:   "cookie:access-token", // "<source>:<name>"
		ErrorHandler:  auth.JWTErrorChecker,
	})

	// routes
	//POST /api/user/login — аутентификация пользователя;
	//POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
	//GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
	//GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
	//POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	//GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.

	e.POST("/api/user/register", controllers.UserRegister())
	e.POST("/api/user/login", controllers.UserLogin())
	e.POST("/api/user/orders", controllers.CreateOrder(), jwtMiddleware)
	e.GET("/api/user/orders", controllers.GetOrders(), jwtMiddleware)
	e.GET("/api/user/balance", controllers.GetBalance(), jwtMiddleware)
	e.POST("/api/user/balance/withdraw", controllers.CreateWithdraw(), jwtMiddleware)
	e.GET("/api/user/withdrawals", controllers.GetWithdrawals(), jwtMiddleware)

	// Start gophermart
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(conf.RunAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the gophermart", err.Error())
		}

		e.Logger.Info("Running gophermart")
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()

	// a timeout of 1 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

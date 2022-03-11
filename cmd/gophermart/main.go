package main

import (
	"Loyalty/configs"
	"Loyalty/internal/client"
	"Loyalty/internal/handler"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	//init logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	logger.SetLevel(logrus.DebugLevel)

	//init configs
	config, err := configs.InitConfig()
	if err != nil {
		logger.Fatal(err)
	}
	//db connection
	db, err := repository.NewDB(context.Background(), config.DatabaseURI)
	if err != nil {
		logger.Fatal("No database connection ")
	}
	//migration
	if err := repository.AutoMigration(viper.GetBool("db.migration.isAllowed"), config.DatabaseURI); err != nil {
		logger.Error("Error: migrations wasn't successful")
	}

	//init accrual client
	c := client.NewAccrualClient(logger, config.AccrualAddress)

	//init main components
	r := repository.NewRepository(db, logger)
	s := service.NewService(r, c, logger)
	h := handler.NewHandler(s, logger)

	//run worker for updating orders queue
	go s.UpdateOrdersQueue()

	//init server
	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: InitRoutes(h),
	}
	//run server
	go server.ListenAndServe()

	logger.Infof("Server started by address: %s", config.ServerAddress)

	//shutdown
	quit := make(chan os.Signal, 1)
	emptyQueue := make(chan struct{})
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	go func() {
		for {
			order := r.TakeFirst()
			if order == "" {
				emptyQueue <- struct{}{}
			}
		}
	}()
	<-emptyQueue
	logrus.Info("Server stopped")
}

func InitRoutes(h *handler.Handler) *gin.Engine {

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

	//orders from user
	user.POST("/orders", h.SaveOrder)
	//withdrawal request
	user.POST("/balance/withdraw", h.Withdraw)
	//getting a list of orders
	user.GET("/orders", h.GetOrders)
	//getting balance
	user.GET("/balance", h.GetBalance)
	//getting information of withdrawals
	user.GET("/balance/withdrawals", h.GetWithdrawals)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Not correct URL"})
	})
	return router
}

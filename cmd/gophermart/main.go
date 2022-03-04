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

	//run accrual server
	// cmd := exec.Command("./cmd/accrual/accrual_linux_amd64")
	//go cmd.Run()

	//mock accrual server
	// go func() {
	// 	time.Sleep(time.Second * 2)
	// 	if err := c.AccrualMock(); err != nil {
	// 		logger.Error(err)
	// 	}
	// }()

	//run worker for updating orders queue
	go s.UpdateOrdersQueue()

	//init server
	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: h.Init(),
	}
	//run server
	go server.ListenAndServe()

	logger.Infof("Server started by address: %s", config.ServerAddress)

	//shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	<-time.After(time.Second * 5)
	logrus.Info("Server stopped")
}

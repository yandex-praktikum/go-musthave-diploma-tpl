package main

import (
	"Loyalty/configs"
	"Loyalty/internal/client"
	"Loyalty/internal/handler"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(new(logrus.TextFormatter))
	logger.SetLevel(logrus.InfoLevel)

	//init configs
	if err := configs.InitConfig(); err != nil {
		logger.Fatal(err)
	}
	//db connection
	db, err := repository.NewDB(context.Background())
	if err != nil {
		logger.Fatal("No database connection ")
	}
	//migration
	if err := repository.AutoMigration(viper.GetBool("db.migration.isAllowed")); err != nil {
		logger.Error("Error: migrations wasn't successful")
	}

	//init main components
	r := repository.NewRepository(db)
	s := service.NewService(r)
	h := handler.NewHandler(s)

	//run accrual server
	cmd := exec.Command("./cmd/accrual/accrual_linux_amd64")
	go cmd.Run()

	//mock accrual server
	accrualClient := client.NewAccrualClient(*logger)
	go func() {
		time.Sleep(time.Second * 2)
		if err := accrualClient.AccrualMock(); err != nil {
			logger.Error(err)
		}
	}()

	//init server
	adr := fmt.Sprint(viper.GetString("host"), ":", viper.GetString("port"))
	server := &http.Server{
		Addr:    adr,
		Handler: h.Init(),
	}
	//run server
	go server.ListenAndServe()

	logger.Infof("Server started by address: %s", adr)

	//shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Info("Server stopped")
}

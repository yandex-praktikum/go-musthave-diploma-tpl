package main

import (
	"Loyalty/configs"
	"Loyalty/internal/handler"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(new(logrus.JSONFormatter))
	logger.SetLevel(logrus.StandardLogger().Level)

	//init configs
	if err := configs.InitConfig(); err != nil {
		logrus.Fatal(err)
	}
	//db connection
	db, err := repository.NewDB(context.Background())
	if err != nil {
		logrus.Fatal("No database connection ")
	}
	//migration
	if err := repository.AutoMigration(viper.GetBool("db.migration.isAllowed")); err != nil {
		logrus.Error("Error: migrations wasn't successful")
	}

	//init main components
	r := repository.NewRepository(db)
	s := service.NewService(r)
	h := handler.NewHandler(s)

	//init server
	adr := fmt.Sprint(viper.GetString("host"), ":", viper.GetString("port"))
	server := &http.Server{
		Addr:    adr,
		Handler: h.Init(),
	}
	//run server
	go server.ListenAndServe()

	logrus.Printf("Server started by address: %s", adr)

	//shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Println("Server stopped")
}

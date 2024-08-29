package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/gophermart/internal/app"
	"github.com/eac0de/gophermart/internal/config"
	"github.com/eac0de/gophermart/pkg/logger"
)

func main() {
	logger.InitLogger("info")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := config.NewConfig()
	app := app.NewApp(ctx, config)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go app.Run()
	<-sigChan
	app.Stop()
}

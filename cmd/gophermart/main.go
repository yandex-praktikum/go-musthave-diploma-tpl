package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/s-lyash/go-musthave-diploma-tpl/internal/app"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	ctx := context.Background()

	a := app.InitApp(ctx)
	a.StartApp(ctx, signals)
}

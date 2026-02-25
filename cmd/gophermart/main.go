package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anon-d/gophermarket/internal/app"
)

func main() {

	app := app.New()
	go func() {
		app.Run()
	}()

	out := make(chan os.Signal, 1)
	signal.Notify(out, syscall.SIGINT, syscall.SIGTERM)
	<-out

	log.Print("Shutdown process is started...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Println("Started force shutdown process")
		os.Exit(1)
	}
}

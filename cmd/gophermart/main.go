package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/Raime-34/gophermart.git/internal/server"
)

func main() {
	logger.InitLogger()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	go server.StartServer(ctx, &wg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	cancel()

	wg.Wait()
	fmt.Println("shutdown gracefully")
}

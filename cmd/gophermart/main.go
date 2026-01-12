package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/config"
	"gophermart/internal/server"
)

func main() {
	cfg := config.Load()

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", cfg.RunAddress)
		if err := srv.ListenAndServe(ctx); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	case <-ctx.Done():
		log.Println("Received shutdown signal, initiating graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		} else {
			log.Println("Server stopped gracefully")
		}
	}
}

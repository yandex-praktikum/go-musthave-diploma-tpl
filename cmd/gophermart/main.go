package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/akashipov/go-musthave-diploma-tpl/internal/server"
	"go.uber.org/zap"
)

func main() {
	server.ParseArgsServer()
	err := server.InitDB()
	if err != nil {
		fmt.Printf("Something wrong with db: %v\n", err)
		return
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Problem with initialization of zap development environment")
		return
	}
	sugar := *logger.Sugar()
	srv := &http.Server{Handler: server.ServerRouter(&sugar)}
	done := make(chan bool)
	go run(srv, done)
	fmt.Println("Server is started")
	r := <-done
	fmt.Println(r)
	fmt.Println("Server is finished...")
}

func run(srv *http.Server, done chan bool) {
	// server.ParseArgsServer()
	// err := server.InitDB()
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigint
		fmt.Println()
		fmt.Printf("Signal: %v\n", sig)
		done <- true
	}()
	srv.Addr = "0.0.0.0:8000"
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

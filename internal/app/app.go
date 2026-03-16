package app

import (
	"context"
	"fmt"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/config"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/handler"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/service"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/store"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	Store   store.Store
	Config  config.Config
	Service service.Service
}

func NewApplication(s store.Store, c config.Config, o service.Service) Application {
	return Application{
		Store:   s,
		Config:  c,
		Service: o,
	}
}

func SetupAPI(s service.Service) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/ping", handler.Test(s)).Methods(http.MethodGet)

	return router

}

func Run(ctx context.Context, c config.Config, router *mux.Router) error {

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	listener, err := net.Listen("tcp", c.RunAddressValue)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %v", c.RunAddressValue, err)
	}

	server := &http.Server{
		Handler:           router,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		if err = server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("Server started and listening on address and port %s", c.RunAddressValue)

	sig := <-cancelChan

	log.Printf("Caught signal %v", sig)
	if err = server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %v", err)
	}

	log.Println("Server shutdown successfully")
	return nil
}

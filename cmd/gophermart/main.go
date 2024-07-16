package main

import (
	"github.com/sashaaro/go-musthave-diploma-tpl/internal"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/adapters"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/handlers"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/infra"
	"log"
	"net/http"
)

func main() {
	internal.InitConfig()

	logger := adapters.CreateLogger()
	pool := infra.CreatePgxPool()
	//nolint:errcheck
	defer pool.Close()

	log.Fatal(http.ListenAndServe(internal.Config.ServerAddress, handlers.CreateServeMux(logger, pool)))
}

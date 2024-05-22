package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/adapters"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/handlers"
	"log"
	"net/http"
)

func main() {
	internal.InitConfig()

	logger := adapters.CreateLogger()
	var pool *pgxpool.Pool

	log.Fatal(http.ListenAndServe(internal.Config.ServerAddress, handlers.CreateServeMux(logger, pool)))
}

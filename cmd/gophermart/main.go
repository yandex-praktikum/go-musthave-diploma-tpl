package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/with0p/gophermart/internal/config"
	"github.com/with0p/gophermart/internal/handlers"
	"github.com/with0p/gophermart/internal/service"
	"github.com/with0p/gophermart/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := config.GetConfig()

	db, dbErr := sql.Open("pgx", config.DataBaseAddress)
	if dbErr != nil {
		fmt.Println(dbErr.Error())
	}
	defer db.Close()

	ctx, cancelInitDB := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelInitDB()

	storage, err := storage.NewStorageDB(ctx, db)
	if err != nil {
		fmt.Println(err.Error())
	}

	service := service.NewServiceGophermart(storage)
	handler := handlers.NewHandlerUserAPI(&service)
	router := handler.GetHandlerUserAPIRouter()

	serverErr := http.ListenAndServe(config.BaseURL, router)
	if err != nil {
		fmt.Println(serverErr.Error())
		return
	}

}

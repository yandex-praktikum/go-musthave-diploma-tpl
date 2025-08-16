package main

import (
	"context"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/app"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/config"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/service"
	"github.com/StarkovPO/go-musthave-diploma-tpl.git/internal/store"
	"log"
	"time"
)

func main() {

	c, err := config.Init()

	if err != nil {
		log.Fatalf("init configuration: %s", err)
	}

	db := store.NewPostgres(store.MustPostgresConnection(c))
	s := service.NewService(db, c)
	application := app.NewApplication(db, c, s)
	router := app.SetupAPI(application.Service)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = app.Run(ctx, application.Config, router)
	if err != nil {
		log.Fatalf("run application: %s", err)
	}

}

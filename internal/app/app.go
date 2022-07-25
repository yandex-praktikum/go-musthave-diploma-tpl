package app

import (
	"github.com/botaevg/gophermart/internal/config"
	"github.com/botaevg/gophermart/internal/handlers"
	"github.com/botaevg/gophermart/internal/repositories"
	"github.com/botaevg/gophermart/internal/service"
	"github.com/botaevg/gophermart/pkg/postgre"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func StartApp() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Print("config error")
		return
	}
	dbpool, err := postgre.NewClient(cfg.DataBaseDsn)
	if err != nil {
		log.Print("DB connect error")
		return
	}

	storage := repositories.NewDB(dbpool)

	auth := service.NewAuth(storage, "")

	r := chi.NewRouter()

	h := handlers.NewHandler(cfg, auth)
	r.Use(middleware.Logger)

	r.Post("/api/user/register", h.RegisterUser)
	r.Post("/api/user/login", h.Login)

	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}

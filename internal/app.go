package server

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/composition"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/config"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type App struct {
	logger logging2.Logger
	router *chi.Mux
	cfg    *config.Config
}

func New(cfg *config.Config, logger logging2.Logger) (*App, error) {

	router := chi.NewRouter()

	logger.Info("Создание userComposite")
	userComposite, err := composition.NewUserComposite(cfg, logger)
	if err != nil {
		logger.Fatalf("Ошибка создания userComposite %v", err)
		return nil, err
	}

	logger.Info("Регистрация /user Роутов")
	userComposite.Handler.Register(router)

	logger.Infof("Start Listen Port %v", *cfg.Port)
	log.Fatal(http.ListenAndServe(*cfg.Port, router))

	app := &App{
		logger: logger,
		router: router,
		cfg:    cfg,
	}

	return app, nil
}

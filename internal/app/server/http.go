package server

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/korol8484/gofermart/internal/app/server/middelewares"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type middlewareHandler func(http.Handler) http.Handler
type registerHandler func(mux *chi.Mux)

type App struct {
	httpServer *http.Server
	cfg        *Config
	log        *zap.Logger

	middlewares []middlewareHandler
	handlers    []registerHandler
}

func NewApp(
	cfg *Config,
	log *zap.Logger,
) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Run(debug bool) error {
	var logBody uint8
	if debug {
		logBody = 1
	}

	router := chi.NewRouter()
	router.Use(
		middleware.Recoverer,
		middlewares.NewLogging(a.log, logBody).LoggingMiddleware,
	)

	for _, m := range a.middlewares {
		router.Use(m)
	}

	a.httpServer = &http.Server{
		Addr:    a.cfg.Listen,
		Handler: router,
	}

	for _, hr := range a.handlers {
		hr(router)
	}

	errCh := make(chan error, 1)

	go func() {
		err := a.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		close(errCh)
	}()

	return <-errCh
}

func (a *App) AddMiddlewares(handler middlewareHandler) {
	a.middlewares = append(a.middlewares, handler)
}

func (a *App) AddHandler(handler registerHandler) {
	a.handlers = append(a.handlers, handler)
}

func (a *App) Stop() {
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	_ = a.httpServer.Shutdown(ctx)
}

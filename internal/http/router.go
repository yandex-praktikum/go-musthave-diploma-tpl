package router

import (
	"context"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/logger"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Config struct {
	Endpoint string
}

type Router struct {
	config Config
}

func New(config Config) *Router {
	return &Router{config}
}

func stub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Stub"))
}

func (router *Router) get(ctx context.Context) chi.Router {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	//r.Use(middleware.NewCompressor(flate.DefaultCompression).Handler)
	//r.Use(dataintergity.NewMiddleware(dataintergity.DataIntegrityMiddlewareConfig{
	//	SigningKey: router.config.SigningKey,
	//}))
	//r.Use(gzipm.GzipMiddleware)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", stub)
		r.Post("/login", stub)

		r.Post("/orders", stub)
		r.Get("/orders", stub)

		r.Get("/balance", stub)
		r.Post("/balance/withdraw", stub)

		r.Post("/withdrawals", stub)
	})

	return r
}

func (router *Router) Run(ctx context.Context) {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get(ctx)))
}

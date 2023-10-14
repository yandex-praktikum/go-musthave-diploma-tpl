package handler

import (
	"github.com/go-chi/chi"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/app/handler/api/user"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/app/middleware"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/auth"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/balance"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/orders"
)

type Handler struct {
	controller  *user.Controller
	middlewares *middleware.MiddleWare
}

func NewHandler(
	authService auth.Auth,
	balanceService balance.Balance,
	orderService orders.Order,
	middlewares *middleware.MiddleWare) *Handler {

	controller := user.NewController(authService, balanceService, orderService)

	return &Handler{
		controller:  controller,
		middlewares: middlewares,
	}
}

func (h Handler) RoutesLink() *chi.Mux {
	r := chi.NewRouter()
	r.Use(h.middlewares.AuthMiddleware)
	r.Use(h.middlewares.GzipMiddleware)
	r.Route("/api", func(r chi.Router) {
		h.controller.User(r)
	})

	return r
}

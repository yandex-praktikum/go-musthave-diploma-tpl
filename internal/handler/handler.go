package handler

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/repository"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/service"
)

type Handler struct {
	authService    service.Autorisation
	ordersService  repository.Orders
	balanceService service.Balance
	cfg            *config.ConfigServer
	log            logger.Logger
}

func NewHandler(auth service.Autorisation, orders repository.Orders, balance service.Balance, cfg *config.ConfigServer, log logger.Logger) *Handler {
	return &Handler{
		authService:    auth,
		ordersService:  orders,
		balanceService: balance,
		cfg:            cfg,
		log:            log,
	}
}

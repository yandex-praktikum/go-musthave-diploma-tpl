package handler

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service"
)

type Handler struct {
	authService    service.Autorisation
	ordersService  service.Orders
	balanceService service.Balance
	cfg            *config.ConfigServer
	log            logger.Logger
}

func NewHandler(auth service.Autorisation, orders service.Orders, balance service.Balance, cfg *config.ConfigServer, log logger.Logger) *Handler {
	return &Handler{
		authService:    auth,
		ordersService:  orders,
		balanceService: balance,
		cfg:            cfg,
		log:            log,
	}
}

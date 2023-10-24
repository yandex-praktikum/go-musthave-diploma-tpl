package handler

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
)

type Handler struct {
	authService    autorisation
	ordersService  orders
	accountService account
	cfg            *config.ConfigServer
	log            logger.Logger
}

func NewHandler(auth autorisation, orders orders, account account, cfg *config.ConfigServer, log logger.Logger) *Handler {
	return &Handler{
		authService:    auth,
		ordersService:  orders,
		accountService: account,
		cfg:            cfg,
		log:            log,
	}
}

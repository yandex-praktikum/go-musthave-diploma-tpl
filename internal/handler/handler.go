package handler

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/service"
)

type Handler struct {
	service *service.Service
	cfg     *config.ConfigServer
	log     logger.Logger
}

func NewHandler(service *service.Service, cfg *config.ConfigServer, log logger.Logger) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
		log:     log,
	}
}

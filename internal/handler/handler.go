package handler

import (
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type Handler struct {
	storage *Storage
	cfg     *config.ConfigServer
	log     logger.Logger
}

func NewHandler(repos *repository.Repository, cfg *config.ConfigServer, log logger.Logger) *Handler {
	return &Handler{
		storage: NewStorage(repos),
		cfg:     cfg,
		log:     log,
	}
}

package handlers

import (
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
	"net/http"
)

type Handler struct {
	service *service.Service
	logs    *logger.Logger
}

func NewHandlers(service *service.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logs:    logger,
	}
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {

}

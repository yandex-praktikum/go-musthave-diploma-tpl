package user

import (
	"context"
	http2 "github.com/GTech1256/go-musthave-diploma-tpl/internal/http"
	logging2 "github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Service interface {
	Ping(ctx context.Context) error
}

type handler struct {
	logger  logging2.Logger
	service Service
}

func NewHandler(logger logging2.Logger, updateService Service) http2.Handler {
	return &handler{
		logger:  logger,
		service: updateService,
	}
}

func (h handler) Register(router *chi.Mux) {
	router.Get("/ping", h.Ping)
}

// Ping /ping/{type}/{name}
func (h handler) Ping(writer http.ResponseWriter, request *http.Request) {
	err := h.service.Ping(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	} else {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok"))
	}
}

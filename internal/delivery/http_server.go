package delivery

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/config"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
}

func NewServerChi(cfg *config.AppConfig, mux *chi.Mux) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Server{
		&http.Server{
			Addr:         cfg.RunAddress,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}, nil
}

func (s *Server) Start() <-chan error {
	serverError := make(chan error, 1)
	go func() {
		zap.L().Info("Starting server", zap.String("addr", s.Server.Addr))
		if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
		close(serverError)
	}()
	return serverError
}

func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	zap.L().Info("Shutting down server", zap.String("addr", s.Server.Addr))
	return s.Server.Shutdown(ctx)
}

package app

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/SmoothWay/gophermart/internal/api"
	"github.com/SmoothWay/gophermart/internal/config"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

func NewServer(l *slog.Logger, addr string) *Server {
	return &Server{
		srv: &http.Server{
			Addr: addr,
		},
		logger: l,
	}
}

func (s *Server) RegisterHandlers(cfg *config.ServerConfig, svc api.Service) {
	swagger, err := api.GetSwagger()
	if err != nil {
		s.logger.Info("GetSwagger error", slog.String("error", err.Error()))
		return
	}

	swagger.Servers = nil

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	r.Use(api.Authenticate([]byte(cfg.Secret)))

	gophermart := api.NewGophermart(s.logger, svc, []byte(cfg.Secret))
	strictHandler := api.NewStrictHandler(gophermart, nil)
	h := api.HandlerFromMux(strictHandler, r)
	s.srv.Handler = h
}

func Run() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.NewServerConfig()

	// db, err := opendb

}

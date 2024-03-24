package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SmoothWay/gophermart/internal/api"
	"github.com/SmoothWay/gophermart/internal/config"
	postgresrepo "github.com/SmoothWay/gophermart/internal/repository/postgres"
	"github.com/SmoothWay/gophermart/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	middleware "github.com/oapi-codegen/nethttp-middleware"
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

	db, err := ConnectDB(cfg.DSN, logger)
	if err != nil {
		logger.Info("DB connection err", slog.String("error", err.Error()))
		return
	}
	defer db.Close()

	repo := postgresrepo.New(db, logger)
	if err != nil {
		logger.Info("DB connection err", slog.String("error", err.Error()))
		return
	}
	client := &http.Client{}
	svc := service.New(logger, repo, client, []byte(cfg.Secret), cfg.AccuralSysAddr)

	server := NewServer(logger, cfg.Host)
	server.RegisterHandlers(cfg, svc)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg = new(sync.WaitGroup)

	wg.Add(1)

	go server.HandleShutdown(ctx, wg)

	orders := repo.ScanOrders(ctx)
	go svc.FetchOrders(ctx, orders)

	logger.Info("Server is listening on", slog.String("address", cfg.Host))
	err = server.srv.ListenAndServe()
	if err != nil {
		logger.Info("Server encountered error", slog.String("error", err.Error()))
		return
	}

	wg.Wait()
}

func ConnectDB(dsn string, l *slog.Logger) (*sql.DB, error) {
	var connection *sql.DB
	var counts int
	var err error
	for {
		connection, err = openDB(dsn)
		if err != nil {
			l.Info("Database not ready...")
			counts++
		} else {
			l.Info("Connected to database")
			break
		}
		if counts > 2 {
			return nil, err
		}
		l.Info(fmt.Sprintf("Retrying to connect after %d seconds\n", counts+2))
		time.Sleep(time.Duration(2+counts) * time.Second)
	}

	instance, err := postgres.WithInstance(connection, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", instance)
	if err != nil {
		return nil, err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return connection, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Server) HandleShutdown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	<-ctx.Done()
	s.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Info("Shutdown server error", slog.String("error", err.Error()))
		return
	}

	s.logger.Info("Server stopped gracefully")
}

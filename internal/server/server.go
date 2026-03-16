package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/accrual"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type Server struct {
	cfg    *config.ConfigServer
	router *gin.Engine
	log    logger.Logger
}

func NewServer(cfg *config.ConfigServer, log logger.Logger) *Server {
	return &Server{
		cfg: cfg,
		log: log,
	}
}

func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db, err := repository.NewPostgresDB(s.cfg.DSN)

	if err != nil {
		s.log.Info("Failed to initialaze db: %s", err.Error())
		panic("Failed to initialaze db")
	}
	s.log.Info("Success connection to db")
	defer db.Close()

	s.router = s.NewRouter(db)
	go func() {
		s.log.Info("Connect listening on port: %s", s.cfg.Port)
		if err := s.router.Run(s.cfg.Port); err != nil {

			s.log.Fatal("Can't ListenAndServe on port", s.cfg.Port)
		}
	}()

	ordersRepo := repository.NewOrdersPostgres(db)
	accrualService := accrual.NewServiceAccrual(ordersRepo, s.log, s.cfg.AccrualPort)

	go accrualService.ProcessedAccrualData(ctx)

	<-ctx.Done()

	return nil

}

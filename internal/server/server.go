package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/accrual"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type server struct {
	cfg    *config.ConfigServer
	router *gin.Engine
	log    logger.Logger
}

func NewServer(cfg *config.ConfigServer, log logger.Logger) *server {
	return &server{
		cfg: cfg,
		log: log,
	}
}

func (s *server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db, err := repository.NewPostgresDB(s.cfg.DSN)

	if err != nil {
		s.log.Info("Failed to initialaze db: %s", err.Error())
	} else {
		s.log.Info("Success connection to db")
		defer db.Close()
	}

	repo := repository.NewRepository(db)
	s.router = s.NewRouter(repo)
	go func() {
		s.log.Info("Connect listening on port: %s", s.cfg.Port)
		if err := s.router.Run(s.cfg.Port); err != nil {

			s.log.Fatal("Can't ListenAndServe on port", s.cfg.Port)
		}
	}()

	as := accrual.NewServiceAccrual(repo, s.log, s.cfg.AccrualPort)

	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			orders, err := as.Storage.Orders.GetOrdersWithStatus()
			if err != nil {
				s.log.Error(err)
			}
			for _, order := range orders {
				ord, err := as.RecieveOrder(ctx, order.Number)
				if err != nil {
					s.log.Error(err)
				}

				err = as.Storage.Orders.ChangeStatusAndSum(ord.Accrual, ord.Status, ord.Number)

				if err != nil {
					s.log.Error(err)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}

}

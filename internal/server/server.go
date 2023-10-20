package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/accrual"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/repository"
	"github.com/tanya-mtv/go-musthave-diploma-tpl/internal/service"
)

type Server struct {
	cfg    *config.ConfigServer
	router *gin.Engine
	log    logger.Logger
	as     *accrual.ServiceAccrual
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
	} else {
		s.log.Info("Success connection to db")
		defer db.Close()
	}

	authRepo := repository.NewAuthPostgres(db)
	ordersRepo := repository.NewOrdersPostgres(db)
	balanceRepo := repository.NewBalancePostgres(db)

	authService := service.NewAuthStorage(authRepo)
	balanceService := service.NewBalanceStorage(balanceRepo)

	s.router = s.NewRouter(authService, ordersRepo, balanceService)
	go func() {
		s.log.Info("Connect listening on port: %s", s.cfg.Port)
		if err := s.router.Run(s.cfg.Port); err != nil {

			s.log.Fatal("Can't ListenAndServe on port", s.cfg.Port)
		}
	}()

	s.as = accrual.NewServiceAccrual(ordersRepo, s.log, s.cfg.AccrualPort)

	go s.ProcessedAccrualData(ctx)

	<-ctx.Done()

	return nil

}

func (s *Server) ProcessedAccrualData(ctx context.Context) {
	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			orders, err := s.as.Storage.GetOrdersWithStatus()
			if err != nil {
				s.log.Error(err)
			}
			for _, order := range orders {
				ord, t, err := s.as.RecieveOrder(ctx, order.Number)
				if err != nil {
					s.log.Error(err)

					if t != 0 {
						time.Sleep(time.Duration(t) * time.Second)
					}
					continue
				}

				err = s.as.Storage.ChangeStatusAndSum(ord.Accrual, ord.Status, ord.Number)

				if err != nil {
					s.log.Error(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}

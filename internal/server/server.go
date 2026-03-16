package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"gophermart/internal/accrual"
	"gophermart/internal/config"
	"gophermart/internal/handler"
	"gophermart/internal/migrations"
	"gophermart/internal/repository"
	"gophermart/internal/service"
)

type Server struct {
	cfg            *config.Config
	db             *sql.DB
	httpServer     *http.Server
	accrualClient  *accrual.Client
	accrualService *service.AccrualService
	stopWorkers    context.CancelFunc
	wg             sync.WaitGroup
}

func New(cfg *config.Config) (*Server, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.Apply(ctx, db); err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)

	jwtService := service.NewJWTService(cfg.JWTSecret)
	authService := service.NewAuthService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, orderRepo, db)

	var accrualClient *accrual.Client
	var accrualService *service.AccrualService

	workersCtx, stopWorkers := context.WithCancel(context.Background())

	if cfg.AccrualSystemAddr != "" {
		cl, err := accrual.New(cfg.AccrualSystemAddr)
		if err != nil {
			stopWorkers()
			return nil, err
		}
		accrualClient = cl
		accrualService = service.NewAccrualService(orderRepo, accrualClient)
	}

	authMiddleware := handler.NewAuthMiddleware(jwtService)

	authHandler := handler.NewAuthHandler(authService, jwtService)
	orderHandler := handler.NewOrderHandler(orderService)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	mux := http.NewServeMux()
	registerRoutes(mux, authMiddleware, authHandler, orderHandler, balanceHandler)

	s := &Server{
		cfg: cfg,
		db:  db,
		httpServer: &http.Server{
			Addr:         cfg.RunAddress,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		accrualClient:  accrualClient,
		accrualService: accrualService,
		stopWorkers:    stopWorkers,
	}

	if accrualService != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			accrualService.StartWorker(workersCtx)
		}()
	}

	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Starting graceful shutdown...")

	if s.stopWorkers != nil {
		s.stopWorkers()
	}

	workerDone := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		log.Println("All workers stopped")
	case <-ctx.Done():
		log.Println("Timeout waiting for workers to stop")
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	if err := s.db.Close(); err != nil {
		log.Printf("Database close error: %v", err)
		return err
	}

	log.Println("Graceful shutdown completed")
	return nil
}

func registerRoutes(
	mux *http.ServeMux,
	authMiddleware *handler.AuthMiddleware,
	authHandler *handler.AuthHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
) {
	mux.HandleFunc("/api/user/register", authHandler.Register)
	mux.HandleFunc("/api/user/login", authHandler.Login)

	mux.HandleFunc("/api/user/orders", authMiddleware.WithAuth(orderHandler.HandleOrders))
	mux.HandleFunc("/api/user/balance", authMiddleware.WithAuth(balanceHandler.GetBalance))
	mux.HandleFunc("/api/user/balance/withdraw", authMiddleware.WithAuth(balanceHandler.Withdraw))
	mux.HandleFunc("/api/user/withdrawals", authMiddleware.WithAuth(balanceHandler.ListWithdrawals))
}

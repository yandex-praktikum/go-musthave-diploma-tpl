package router

import (
	"github.com/EestiChameleon/GOphermart/internal/app/cfg"
	h "github.com/EestiChameleon/GOphermart/internal/app/router/handlers"
	"github.com/EestiChameleon/GOphermart/internal/app/router/mw"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"time"
)

func Start() error {
	// Chi instance
	router := chi.NewRouter()

	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Routes
	router.With(mw.AuthCheck).Get("/api/user/orders", h.UserOrdersList)
	router.With(mw.AuthCheck).Get("/api/user/balance", h.UserBalance)
	router.With(mw.AuthCheck).Get("/api/user/balance/withdrawals", h.UserBalanceWithdrawals)

	router.Post("/api/user/register", h.UserRegister)
	router.Post("/api/user/login", h.UserLogin)

	router.With(mw.AuthCheck).Post("/api/user/orders", h.UserAddOrder)
	router.With(mw.AuthCheck).Post("/api/user/balance/withdraw", h.UserBalanceWithdraw)

	// Start server
	s := http.Server{
		Addr:    cfg.Envs.RunAddr,
		Handler: router,
		// ReadTimeout: 30 * time.Second, // customize http.Server timeouts
	}

	log.Println("SERVER STARTED AT", time.Now().Format(time.RFC3339))
	return s.ListenAndServe()
}

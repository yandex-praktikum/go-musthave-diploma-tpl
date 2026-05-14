package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/errors"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/handler/middleware"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/service"
)

var _ iService = &service.Service{}

type iService interface {
	Register(ctx context.Context, reg *model.LoginPassword) (string, error)
	Login(ctx context.Context, reg *model.LoginPassword) (string, error)

	ConfirmOrder(ctx context.Context, order string) (bool, error)
	GetOrdersOfUser(ctx context.Context) ([]model.Order, error)
	GetBalance(ctx context.Context) (model.Balance, error)
	Withdrawn(ctx context.Context, req *model.BalanceWithdraw) error
	GetWithdrawns(ctx context.Context) ([]model.Withdraw, error)
}

type Handler struct {
	cfg *Config
	chi.Router

	service iService
}

func NewHandler(cfg *Config, service iService) *Handler {
	h := &Handler{
		cfg:     cfg,
		service: service,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(chimiddleware.StripSlashes)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", interceptError(h.handleRegister))
		r.Post("/login", interceptError(h.handleLogin))

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.EncryptionKey))

			r.Post("/orders", interceptError(h.handlePostOrders))
			r.Get("/orders", interceptError(h.handleGetOrders))
			r.Get("/withdrawals", interceptError(h.handleWithdraw))
			r.Get("/balance", interceptError(h.handleBalance))
			r.Post("/balance/withdraw", interceptError(h.handleBalanceWithdraw))
		})
	})

	h.Router = r
	return h
}

// Перехватчик кастомных ошибок и отправка в http формате
func interceptError(hander func(w http.ResponseWriter, r *http.Request) *errors.AppError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := hander(w, r)
		if err != nil {
			SendJSONResponse(w, err.Msg, err.Code)
			logger.Log.Info(err.Log)
		}
	}
}

// отправка ответа с соответсвующими полями
func SendJSONResponse(w http.ResponseWriter, message interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(message)
}

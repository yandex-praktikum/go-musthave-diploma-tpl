package app

import (
	"github.com/rs/zerolog"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	"github.com/k-morozov/go-musthave-diploma-tpl/core"
	"github.com/k-morozov/go-musthave-diploma-tpl/servers"
)

type App struct {
	core   core.Core
	server servers.Server
	config config.ServiceConfig
}

func NewApp(cfg config.ServiceConfig, log zerolog.Logger, opts ...servers.ServiceOption) (App, error) {
	db, err := store.NewPostgresStore(cfg.DatabaseDsn)
	if err != nil {
		return App{}, err
	}
	c := core.NewBasicCore(db, log)
	s := servers.NewServer(opts...)

	s.AddHandler(servers.Post, "/api/user/register", c.Register)
	s.AddHandler(servers.Post, "/api/user/login", c.Login)
	s.AddHandler(servers.Post, "/api/user/orders", c.AddOrder)
	s.AddHandler(servers.Get, "/api/user/orders", c.GetOrders)
	s.AddHandler(servers.Get, "/api/user/balance", c.GetUserBalance)
	s.AddHandler(servers.Post, "/api/user/balance/withdraw", c.Withdraw)
	s.AddHandler(servers.Get, "/api/user/withdrawals", c.Withdrawals)

	return App{
		core:   c,
		server: s,
		config: cfg,
	}, nil
}

func (a *App) Run() {
	if err := a.server.Start(a.config.ServiceDomain); err != nil {
		// @TODO
	}
}

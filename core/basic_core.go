package core

import (
	"github.com/k-morozov/go-musthave-diploma-tpl/clients"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/actions"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/process"
)

type BasicCore struct {
	store store.Store
	proc  process.Process
}

var _ Core = &BasicCore{}

func NewBasicCore(store store.Store, log zerolog.Logger) Core {

	// @TODO options
	p := process.NewTimerProcess(store, clients.NewLoyaltyPointsCalculationSystem(), log)
	p.Run()

	return &BasicCore{
		store: store,
		proc:  p,
	}
}

func (c *BasicCore) Register(rw http.ResponseWriter, req *http.Request) {
	actions.Register(rw, req, c.store)
}

func (c *BasicCore) Login(rw http.ResponseWriter, req *http.Request) {
	actions.Login(rw, req, c.store)
}

func (c *BasicCore) AddOrder(rw http.ResponseWriter, req *http.Request) {
	actions.AddOrder(rw, req, c.store)
}

func (c *BasicCore) GetOrders(rw http.ResponseWriter, req *http.Request) {
	actions.GetOrders(rw, req, c.store)
}

func (c *BasicCore) GetUserBalance(rw http.ResponseWriter, req *http.Request) {
	actions.GetUserBalance(rw, req, c.store)
}

func (c *BasicCore) Withdraw(rw http.ResponseWriter, req *http.Request) {
	actions.Withdraw(rw, req, c.store)
}

func (c *BasicCore) Withdrawals(rw http.ResponseWriter, req *http.Request) {
	actions.Withdrawals(rw, req, c.store)
}

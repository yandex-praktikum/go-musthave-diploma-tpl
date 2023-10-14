package user

import (
	"github.com/go-chi/chi"

	user "github.com/s-lyash/go-musthave-diploma-tpl/internal/app/handler/api/user/auth"
	wallet "github.com/s-lyash/go-musthave-diploma-tpl/internal/app/handler/api/user/balance"
	indent "github.com/s-lyash/go-musthave-diploma-tpl/internal/app/handler/api/user/order"

	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/auth"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/balance"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/orders"
)

type Controller struct {
	user    *user.Handler
	balance *wallet.Handler
	order   *indent.Handler
}

func NewController(auth auth.Auth, balance balance.Balance, orders orders.Order) *Controller {
	balanceHandler := wallet.NewHandler(balance)
	orderHandler := indent.NewHandler(orders)
	authHandler := user.NewHandler(auth)

	return &Controller{
		user:    authHandler,
		balance: balanceHandler,
		order:   orderHandler,
	}
}

func (ctr Controller) User(r chi.Router) {
	r.Route("/user", func(r chi.Router) {
		ctr.user.AuthRouter(r)
		ctr.balance.Balance(r)
		ctr.order.Order(r)
	})
}

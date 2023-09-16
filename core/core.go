package core

import "net/http"

type Core interface {
	Register(rw http.ResponseWriter, req *http.Request)
	Login(rw http.ResponseWriter, req *http.Request)
	AddOrder(rw http.ResponseWriter, req *http.Request)
	GetOrders(rw http.ResponseWriter, req *http.Request)
	GetUserBalance(rw http.ResponseWriter, req *http.Request)
	Withdraw(rw http.ResponseWriter, req *http.Request)
	Withdrawals(rw http.ResponseWriter, req *http.Request)
}

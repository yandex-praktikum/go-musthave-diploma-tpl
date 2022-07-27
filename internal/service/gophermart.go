package service

import (
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/repositories"
)

type Gophermart struct {
	storage repositories.Storage
}

func NewGophermart(storage repositories.Storage) Gophermart {
	return Gophermart{storage: storage}
}

func (g Gophermart) CheckOrder(number uint) (uint, error) {
	return g.storage.CheckOrder(number)
}

func (g Gophermart) AddOrder(number uint, userID uint) error {
	return g.storage.AddOrder(number, userID)
}

func (g Gophermart) GetListOrders(userid uint) ([]models.OrderAPI, error) {
	ListOrders, err := g.storage.GetListOrders(userid)
	if err != nil {
		return nil, err
	}
	var ListOrdersAPI []models.OrderAPI
	for _, v := range ListOrders {
		x := models.OrderAPI{}
		x.Number = v.OrderNumber
		x.Status = v.Status
		//x.Accrual = v.Accrual
		x.UploadedAt = v.Date
		ListOrdersAPI = append(ListOrdersAPI, x)
	}
	return ListOrdersAPI, err
}

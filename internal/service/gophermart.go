package service

import (
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/repositories"
	"log"
)

type Gophermart struct {
	storage repositories.Storage
}

func NewGophermart(storage repositories.Storage) Gophermart {
	return Gophermart{storage: storage}
}

func (g Gophermart) CheckOrder(number string) (uint, error) {
	return g.storage.CheckOrder(number)
}

func (g Gophermart) AddOrder(number string, userID uint) error {
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
		x.Accrual = v.Accrual
		x.UploadedAt = v.Date
		ListOrdersAPI = append(ListOrdersAPI, x)
	}
	return ListOrdersAPI, err
}

func (g Gophermart) BalanceUser(userID uint) (models.AccountBalanceAPI, error) {
	current, err := g.storage.BalanceUser(userID)
	if err != nil {
		return models.AccountBalanceAPI{}, err
	}
	withdrawn, err := g.storage.SumWithdrawn(userID)
	if err != nil {
		return models.AccountBalanceAPI{}, err
	}
	return models.AccountBalanceAPI{
		Current:   current,
		Withdrawn: withdrawn,
	}, err
}

func (g Gophermart) WithdrawRequest(withdrawnreq models.Withdraw, userID uint) (bool, error) {

	balance, err := g.storage.BalanceUser(userID)
	if err != nil {
		log.Print(err)
		return false, err
	}
	if withdrawnreq.Sum > balance {
		log.Print("sum > balance")
		return false, err
	}

	err = g.storage.ChangeBalance(models.AccountBalance{
		UserID:      userID,
		OrderNumber: withdrawnreq.Order,
		TypeMove:    "withdraw",
		SumAccrual:  withdrawnreq.Sum,
		Balance:     balance - withdrawnreq.Sum,
	})
	if err != nil {
		log.Print(err)
		return false, err
	}
	return true, nil
}

func (g Gophermart) ListWithdraw(userid uint) ([]models.Withdraw, error) {
	return g.storage.ListWithdraw(userid)
}

func (g Gophermart) UpdateOrders(order models.OrderES) error {
	log.Print("g UpdateOrders")
	return g.storage.UpdateOrders(order)
}

func (g Gophermart) AccrualRequest(order models.OrderES) error {
	log.Print("g AccrualRequest")
	userID, err := g.storage.OwnerOrders(order.Order)
	if err != nil {
		log.Print(err)
		return err
	}
	balance, err := g.storage.BalanceUser(userID)
	if err != nil {
		log.Print(err)
		return err
	}

	err = g.storage.ChangeBalance(models.AccountBalance{
		UserID:      userID,
		OrderNumber: order.Order,
		TypeMove:    "accrual",
		SumAccrual:  order.Accrual,
		Balance:     balance + order.Accrual,
	})
	if err != nil {
		log.Print(err)
		return err
	}
	return err
}

package service

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tartushkin/go-musthave-diploma-tpl.git/internal/model"
)

func (s *Short) GetOrderList(log string) []*model.UserOrderRes {
	user, _ := s.UserCH[log]
	orders := []*model.UserOrderRes{}

	if len(user.OrderList) < 1 {
		return orders
	}

	sort.Slice(orders, func(i, j int) bool {
		return user.OrderList[i].Created.Before(user.OrderList[j].Created)
	})

	for _, order := range user.OrderList {
		orders = append(orders, &model.UserOrderRes{
			Order:   order.OrderID,
			Status:  order.Status,
			Accrual: order.Accural,
			Created: order.Created.Format(time.RFC3339),
		})

	}

	return orders
}

func (s *Short) WithdrawnBalance(log, order string, amount decimal.Decimal) error {
	err := s.Repo.WithdrawnBalance(s.Ctx, log, order, amount)
	if err != nil {
		return err
	}
	return nil
}

func (s *Short) GetInfoWithdrawnBalance(log string) ([]*model.WithdrawReq, error) {
	list, err := s.Repo.GetInfoWithdrawnBalance(s.Ctx, log)
	if err != nil {
		return nil, err
	}
	return list, err
}

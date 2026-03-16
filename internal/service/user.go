package service

import (
	"context"
	"fmt"
	"musthave/internal/model"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

func (m *Market) LoadStorageUser(ctx context.Context) error {
	list, err := m.Repo.GetUserList(ctx)
	if err != nil {
		return err
	}

	for _, user := range list {
		orderList, err := m.Repo.GetOrderList(ctx, user.Login)
		if err != nil {
			return err
		}

		user.OrderList = make(map[int]*model.Order)
		for _, order := range orderList {
			user.OrderList[order.OrderID] = order
		}
		m.Mu.Lock()
		m.UserCH[user.Login] = user
		m.Mu.Unlock()
	}
	return nil
}

func (m *Market) RegisterUser(log string, pass string) error {
	m.Lg.Info("RegisterUser.start - начало регестрации нового пользователя: " + log)

	err := m.create(log, pass) // упаковка пользователя
	if err != nil {
		return err
	}

	return nil
}

// create - проверка и регистрация нового пользователя
func (m *Market) create(log string, pass string) error {
	m.Mu.RLock()
	_, ok := m.UserCH[log]
	m.Mu.RUnlock()
	if ok {
		return fmt.Errorf(" пользователь с логином %s уже существует", log) // логин есть, нужна кастомная ошибка и проверка на каастом извне
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	user := &model.User{
		Login:     log,
		PassHash:  string(hash),
		OrderList: make(map[int]*model.Order),
	}
	err := m.Repo.RegisterUser(m.Ctx, log, string(hash))
	if err != nil {
		return err
	}
	m.Lg.Info("RegisterUser.progress - успешно добавли нового пользователя в БД")

	m.Mu.RLock()
	m.UserCH[user.Login] = user
	m.Mu.RUnlock()
	m.Lg.Info("RegisterUser.progress - добавли нового пользователя в кеш ")

	return nil
}

func (m *Market) GetMyBalance(login string) (decimal.Decimal, decimal.Decimal, error) {
	m.Lg.Info("GetMyBanance.start - подсчет баланса для пользователя: " + login)
	cb, tw, err := m.Repo.GetInfoMyBalance(m.Ctx, login)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	m.Lg.Info(fmt.Sprintf("GetMyBanance.start - подсчет баланса для пользователя: %s - завершен. Баланс: %v, Выведено: %v", login, cb, tw))

	return cb, tw, nil
}

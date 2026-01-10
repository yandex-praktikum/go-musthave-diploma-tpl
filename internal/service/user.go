package service

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tartushkin/go-musthave-diploma-tpl.git/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func (s *Short) RegisterUser(log string, pass string) error {
	s.Lg.Info("RegisterUser.start - начало регестрации нового пользователя")

	err := s.create(log, pass) // упаковка пользователя
	if err != nil {
		return err
	}

	return nil
}

// create - проверка и регистрация нового пользователя
func (s *Short) create(log string, pass string) error {
	_, ok := s.UserCH[log]
	if ok {
		return fmt.Errorf(" пользователь с логином %s уже существует", log) // логин есть, нужна кастомная ошибка и проверка на каастом извне
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	user := &model.User{
		Login:    log,
		PassHash: string(hash),
	}
	err := s.Repo.RegisterUser(s.Ctx, log, pass)
	if err != nil {
		return err
	}
	s.Lg.Info("RegisterUser.progress - успешно добавли нового пользователя в БД")

	s.mu.RLock()
	s.UserCH[user.Login] = user
	s.mu.RUnlock()
	s.Lg.Info("RegisterUser.progress - добавли нового пользователя в кеш ")

	return nil
}

func (s *Short) GetMyBalance(login string) (decimal.Decimal, decimal.Decimal, error) {
	s.Lg.Info("GetMyBanance.start - подсчет баланса для пользователя: " + login)
	cb, tw, err := s.Repo.GetInfoMyBalance(s.Ctx, login)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	s.Lg.Info(fmt.Sprintf("GetMyBanance.start - подсчет баланса для пользователя: %s - завершен. Баланс: %v, Выведено: %v", login, cb, tw))

	return cb, tw, nil
}

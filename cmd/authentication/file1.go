package authentication

import (
	"crypto/md5"
	"encoding/hex"
)

type Name struct {
	Login           string `json:"-"`
	Client          string `json:"-"`
	CurrentPoints   int    `json:"current"`
	WithdrawnPoints int    `json:"withdrawn"`
	Coockie         string `json:"-"`
}

func New() *Name {
	return &Name{}
}

func func1(login string, pass string) (user Name, err error) {
	user.Login = login
	logPas := []byte(login + pass)
	h := md5.New()
	h.Write(logPas)
	userByte := h.Sum(nil)
	user.Client = hex.EncodeToString(userByte) // возвращать для хранения в таблице как пользователя
	userByteCrypt, err := CryptoToken(userByte)
	if err != nil {
		return Name{}, err
	}
	user.Coockie = hex.EncodeToString(userByteCrypt) // возвращать для передачи клиенту (временная кука)
	return user, nil
}

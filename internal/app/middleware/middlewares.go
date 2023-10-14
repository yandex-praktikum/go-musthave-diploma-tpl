package middleware

import (
	"github.com/s-lyash/go-musthave-diploma-tpl/config"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/auth"
)

type MiddleWare struct {
	Key  models.IDKey
	auth auth.Auth
	conf *config.Config
}

func New(auth auth.Auth, conf *config.Config) *MiddleWare {
	return &MiddleWare{
		auth: auth,
		conf: conf,
	}
}

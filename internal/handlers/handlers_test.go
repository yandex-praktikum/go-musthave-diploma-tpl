package handlers

import (
	"github.com/sashaaro/go-musthave-diploma-tpl/internal"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/adapters"
	"net/http"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	httpClient := http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	internal.InitConfig()

	logger := adapters.CreateLogger()
	_ = logger
	_ = httpClient

	t.Run("create short url, pass through short url", func(t *testing.T) {

	})

}

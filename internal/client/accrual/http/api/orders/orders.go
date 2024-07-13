package orders

import (
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/http/client"
	"github.com/GTech1256/go-musthave-diploma-tpl/pkg/logging"
	//"github.com/GTech1256/go-yandex-metrics-tpl/internal/agent/client/server/http/client"
	//"github.com/GTech1256/go-yandex-metrics-tpl/pkg/logging"
)

type update struct {
	HTTPClient client.ClientHTTP
	BaseURL    string
	logger     logging.Logger
}

func New(HTTPClient client.ClientHTTP, BaseURL string, logger logging.Logger) *update {
	return &update{
		HTTPClient: HTTPClient,
		BaseURL:    BaseURL,
		logger:     logger,
	}
}

package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
)

const (
	RetryMax     int           = 3
	RetryWaitMin time.Duration = 1 * time.Second
	RetryMedium  time.Duration = 3 * time.Second
	RetryWaitMax time.Duration = 5 * time.Second
)

type ServiceAccrual struct {
	Storage    repository.Orders
	httpClient *http.Client
	log        logger.Logger
	addr       string
}

func NewServiceAccrual(stor repository.Orders, log logger.Logger, addr string) *ServiceAccrual {
	return &ServiceAccrual{
		Storage:    stor,
		httpClient: &http.Client{},
		log:        log,
		addr:       addr,
	}
}

func (s *ServiceAccrual) RecieveOrder(ctx context.Context, number string) (models.OrderResponse, error, int) {
	var orderResp models.OrderResponse
	url := fmt.Sprintf("%s/api/orders/%s", s.addr, number)
	s.log.Info("Recieving order from accrual system ", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.log.Error(err)
		return orderResp, err, 0
	}

	resp, err := s.httpClient.Do(req)

	if err != nil {
		s.log.Debug("Can't get message")
		return orderResp, err, 0
	}
	defer resp.Body.Close()

	s.log.Info("Get response status ", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:

		jsonData, err := io.ReadAll(resp.Body)
		if err != nil {
			s.log.Error(err)
			return orderResp, err, 0
		}

		if err := json.Unmarshal(jsonData, &orderResp); err != nil {
			s.log.Error(err)
			return orderResp, err, 0
		}
		s.log.Info("Get data from accrual system  ", orderResp)

		if orderResp.Status == "REGISTERED" {
			orderResp.Status = "NEW"
		}
		s.log.Info("Get data", orderResp)
		return orderResp, nil, 0
	case http.StatusNoContent:
		s.log.Info("No content in request ")
		return orderResp, errors.New("NoContent"), 0
	case http.StatusTooManyRequests:
		s.log.Info("Too Many Requests ")

		retryHeder := resp.Header.Get("Retry-After")
		retryafter, err := strconv.Atoi(retryHeder)
		if err != nil {
			return orderResp, errors.New("TooManyRequests"), 0
		}

		return orderResp, errors.New("TooManyRequests"), retryafter
	}
	return orderResp, nil, 0
}

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

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
)

const (
	RetryMax     int           = 3
	RetryWaitMin time.Duration = 1 * time.Second
	RetryMedium  time.Duration = 3 * time.Second
	RetryWaitMax time.Duration = 5 * time.Second
)

type ServiceAccrual struct {
	Storage    orders
	httpClient *http.Client
	log        logger.Logger
	addr       string
}

func NewServiceAccrual(stor orders, log logger.Logger, addr string) *ServiceAccrual {
	return &ServiceAccrual{
		Storage:    stor,
		httpClient: &http.Client{},
		log:        log,
		addr:       addr,
	}
}

func (s *ServiceAccrual) ProcessedAccrualData(ctx context.Context) {
	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			orders, err := s.Storage.GetOrdersWithStatus()
			if err != nil {
				s.log.Error(err)
			}
			for _, order := range orders {
				ord, t, err := s.RecieveOrder(ctx, order.Number)
				if err != nil {
					s.log.Error(err)

					if t != 0 {
						time.Sleep(time.Duration(t) * time.Second)
					}
					continue
				}

				err = s.Storage.ChangeStatusAndSum(ord.Accrual, ord.Status, ord.Number)

				if err != nil {
					s.log.Error(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}

func (s *ServiceAccrual) RecieveOrder(ctx context.Context, number string) (models.OrderResponse, int, error) {
	var orderResp models.OrderResponse
	url := fmt.Sprintf("%s/api/orders/%s", s.addr, number)
	s.log.Info("Recieving order from accrual system ", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.log.Error(err)
		return orderResp, 0, err
	}

	resp, err := s.httpClient.Do(req)

	if err != nil {
		s.log.Debug("Can't get message")
		return orderResp, 0, err
	}
	defer resp.Body.Close()

	s.log.Info("Get response status ", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:

		jsonData, err := io.ReadAll(resp.Body)
		if err != nil {
			s.log.Error(err)
			return orderResp, 0, err
		}

		if err := json.Unmarshal(jsonData, &orderResp); err != nil {
			s.log.Error(err)
			return orderResp, 0, err
		}
		s.log.Info("Get data from accrual system  ", orderResp)

		if orderResp.Status == "REGISTERED" {
			orderResp.Status = "NEW"
		}
		s.log.Info("Get data", orderResp)
		return orderResp, 0, nil
	case http.StatusNoContent:
		s.log.Info("No content in request ")
		return orderResp, 0, errors.New("NoContent")
	case http.StatusTooManyRequests:
		s.log.Info("Too Many Requests ")

		retryHeder := resp.Header.Get("Retry-After")
		retryafter, err := strconv.Atoi(retryHeder)
		if err != nil {
			return orderResp, 0, errors.New("TooManyRequests")
		}

		return orderResp, retryafter, errors.New("TooManyRequests")
	}
	return orderResp, 0, nil
}

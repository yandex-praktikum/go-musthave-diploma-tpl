package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"

	"github.com/hashicorp/go-retryablehttp"
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
	httpClient *retryablehttp.Client
	log        logger.Logger
	addr       string
}

func NewServiceAccrual(stor repository.Orders, log logger.Logger, addr string) *ServiceAccrual {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = RetryMax
	retryClient.RetryWaitMin = RetryWaitMin
	retryClient.RetryWaitMax = RetryWaitMax
	retryClient.Backoff = backoff

	return &ServiceAccrual{
		Storage:    stor,
		httpClient: retryClient,
		log:        log,
		addr:       addr,
	}
}

func (s *ServiceAccrual) RecieveOrder(ctx context.Context, number string) (models.OrderResponse, error) {
	var orderResp models.OrderResponse
	url := fmt.Sprintf("%s/api/orders/%s", s.addr, number)
	s.log.Info("Recieving order from accrual system ", url)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.log.Error(err)
		return orderResp, err
	}

	resp, err := s.httpClient.Do(req)

	if err != nil {
		s.log.Debug("Can't get message")
		return orderResp, err
	}
	defer resp.Body.Close()

	s.log.Info("Get response status ", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:

		jsonData, err := io.ReadAll(resp.Body)
		if err != nil {
			s.log.Error(err)
			return orderResp, err
		}

		if err := json.Unmarshal(jsonData, &orderResp); err != nil {
			s.log.Error(err)
			return orderResp, err
		}
		s.log.Info("Get data from accrual system  ", orderResp)

		if orderResp.Status == "REGISTERED" {
			orderResp.Status = "NEW"
		}
		s.log.Info("Get data", orderResp)
		return orderResp, nil
	case http.StatusNoContent:
		s.log.Info("No content in request ")
		return orderResp, errors.New("NoContent")
	case http.StatusTooManyRequests:
		s.log.Info("Too Many Requests ")
		return orderResp, errors.New("TooManyRequests")
	}
	return orderResp, nil
}

func backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}

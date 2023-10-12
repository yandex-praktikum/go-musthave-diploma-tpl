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
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/constants"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
)

type ServiceAccrual struct {
	Storage    *repository.Repository
	httpClient *retryablehttp.Client
	log        logger.Logger
	addr       string
}

func NewServiceAccrual(stor *repository.Repository, log logger.Logger, addr string) *ServiceAccrual {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = constants.RetryMax
	retryClient.RetryWaitMin = constants.RetryWaitMin
	retryClient.RetryWaitMax = constants.RetryWaitMax
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
	url := fmt.Sprintf("http://%s/api/order/", s.addr)
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

	switch resp.StatusCode {
	case http.StatusOK:

		jsonData, err := io.ReadAll(resp.Body)
		if err != nil {
			s.log.Error(err)
			return orderResp, err
		}
		defer resp.Body.Close()

		if err := json.Unmarshal(jsonData, &orderResp); err != nil {
			s.log.Error(err)
			return orderResp, err
		}
		if orderResp.Status == "REGISTERED" {
			orderResp.Status = "NEW"
		}

		s.log.Info("Get data", orderResp)
		return orderResp, nil
	case http.StatusNoContent:
		return orderResp, errors.New("NoContent")
	case http.StatusTooManyRequests:
		return orderResp, errors.New("TooManyRequests")
	}
	return orderResp, nil

}

func backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}

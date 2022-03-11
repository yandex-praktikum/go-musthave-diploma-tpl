package client

import (
	"Loyalty/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

const StatusRegistered string = "REGISTERED"
const StatusInvalid string = "INVALID"
const StatusProcessing string = "PROCESSING"
const StatusProcessed string = "PROCESSED"

type AccrualClient struct {
	client  *http.Client
	logger  *logrus.Logger
	address string
}

func NewAccrualClient(logger *logrus.Logger, address string) *AccrualClient {
	return &AccrualClient{
		client:  &http.Client{},
		logger:  logger,
		address: address,
	}
}

func (c *AccrualClient) SentOrder(order string) (*models.Accrual, error) {
	url := fmt.Sprint(c.address, "/api/orders/", order)
	resp, err := c.client.Get(url)
	c.logger.Infof("Accrual response: %v", resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var accrual models.Accrual
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		err := errors.New("error: not found result")
		return nil, fmt.Errorf(`%w`, err)
	}
	if err := json.Unmarshal(body, &accrual); err != nil {
		return nil, err
	}
	c.logger.Infof("Accrual request: %s, response: %v", order, accrual)

	return &accrual, nil
}

package client

import (
	"Loyalty/internal/models"
	"bytes"
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

func NewAccrualClient(logger logrus.Logger, address string) *AccrualClient {
	return &AccrualClient{
		client:  &http.Client{},
		logger:  &logger,
		address: address,
	}
}

type cashback struct {
	Mutch      string `json:"match"`
	Reward     int    `json:"reward"`
	RewardType string `json:"reward_type"`
}

type goods struct {
	Description string  `json:"description"`
	Price       float32 `json:"price"`
}

type order struct {
	Number string  `json:"order"`
	Goods  []goods `json:"goods"`
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

func (c *AccrualClient) AccrualMock() error {
	cashbacks := []cashback{
		{
			Mutch:      "IPhone",
			Reward:     10,
			RewardType: "%",
		},
		{
			Mutch:      "Samsung",
			Reward:     5,
			RewardType: "%",
		},
		{
			Mutch:      "Huawei",
			Reward:     20,
			RewardType: "%",
		},
	}
	orders := []order{
		{
			Number: "123455",
			Goods: []goods{
				{
					Description: "IPhone 11",
					Price:       75000.0,
				},
			},
		},
		{
			Number: "21312541",
			Goods: []goods{
				{
					Description: "Samsung A51",
					Price:       33000.0,
				},
			},
		},
		{
			Number: "7643497",
			Goods: []goods{
				{
					Description: "Huawei P30Lite",
					Price:       17598.5,
				},
			},
		},
	}
	url := fmt.Sprint(c.address, "/api/goods")

	for _, val := range cashbacks {
		body, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buffer := bytes.NewBuffer(body)
		resp, err := c.client.Post(url, "application/json", buffer)
		if err != nil {
			return err
		}
		c.logger.Infof("Accrual mock. Request: %v, status: %s", string(body), resp.Status)
	}
	url = fmt.Sprint(c.address, "/api/orders")
	for _, val := range orders {
		body, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buffer := bytes.NewBuffer(body)
		resp, err := c.client.Post(url, "application/json", buffer)
		if err != nil {
			return err
		}
		c.logger.Infof("Accrual mock. Request: %v, status: %s", string(body), resp.Status)
	}

	return nil
}

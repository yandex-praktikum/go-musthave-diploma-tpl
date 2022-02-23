package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type client struct {
	logger *logrus.Logger
}

func NewAccrualClient(logger logrus.Logger) *client {
	return &client{logger: &logger}
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

func (c *client) SentOrder(order string) error {
	url := fmt.Sprint("http://localhost:8080/api/orders/", order)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) AccrualMock() error {
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
	url := "http://localhost:8080/api/goods"

	for _, val := range cashbacks {
		body, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buffer := bytes.NewBuffer(body)
		resp, err := http.Post(url, "application/json", buffer)
		if err != nil {
			return err
		}
		c.logger.Infof("Request: %v, status: %s", string(body), resp.Status)
	}
	url = "http://localhost:8080/api/orders"
	for _, val := range orders {
		body, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buffer := bytes.NewBuffer(body)
		resp, err := http.Post(url, "application/json", buffer)
		if err != nil {
			return err
		}
		c.logger.Infof("Request: %v, status: %s", string(body), resp.Status)
	}

	return nil
}

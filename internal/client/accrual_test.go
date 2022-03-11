package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type cacheback struct {
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

func (c *AccrualClient) AccrualMock() error {
	cashbacks := []cacheback{
		{
			Mutch:      "IPhone",
			Reward:     7,
			RewardType: "%",
		},
		{
			Mutch:      "Samsung",
			Reward:     17,
			RewardType: "%",
		},
		{
			Mutch:      "Huawei",
			Reward:     12,
			RewardType: "%",
		},
	}
	orders := []order{
		{
			Number: "123455",
			Goods: []goods{
				{
					Description: "IPhone 11",
					Price:       75000.74,
				},
			},
		},
		{
			Number: "21312541",
			Goods: []goods{
				{
					Description: "Samsung A51",
					Price:       33000.47,
				},
			},
		},
		{
			Number: "7643497",
			Goods: []goods{
				{
					Description: "Huawei P30Lite",
					Price:       17598.53,
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
		defer resp.Body.Close()
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
		defer resp.Body.Close()
		c.logger.Infof("Accrual mock. Request: %v, status: %s", string(body), resp.Status)
	}

	return nil
}

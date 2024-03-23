package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/A-Kuklin/gophermart/internal/domain/entities"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/accrual/types"
)

const (
	groupAPI  = "/api/orders/"
	batchSize = 10
	rateLimit = 3
)

type Storage interface {
	GetOrders(ctx context.Context, limit int) ([]types.Order, error)
	UpdateOrderAccrual(ctx context.Context, orderAccrual types.AccrualResponse) error
}

type Client struct {
	storage             Storage
	client              *http.Client
	host                string
	logger              logrus.FieldLogger
	chProcessingOrders  chan types.Order
	chProcessingAccrual chan types.AccrualResponse
}

func NewClient(storage Storage, host string, logger logrus.FieldLogger) *Client {
	return &Client{
		storage:             storage,
		client:              &http.Client{},
		host:                host,
		logger:              logger,
		chProcessingOrders:  make(chan types.Order),
		chProcessingAccrual: make(chan types.AccrualResponse),
	}
}

func (c *Client) Run(ctx context.Context) {
	go c.GetOrders(ctx)
	go c.GetAccrual(ctx)
	go c.processAccrual(ctx)
}

func (c *Client) GetOrders(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			orders, err := c.storage.GetOrders(ctx, batchSize)
			if err != nil {
				c.logger.WithError(err).Error("GetOrders for accrual client error")
			}

			for _, order := range orders {
				c.chProcessingOrders <- order
			}
		}
	}
}

func (c *Client) GetAccrual(ctx context.Context) {
	var wg sync.WaitGroup

	for rl := 0; rl < rateLimit; rl++ {
		select {
		case order := <-c.chProcessingOrders:
			wg.Add(1)
			c.logger.Debugf("Accrual worker %d starting", rl)
			go c.getAccrualSum(&wg, order)
		case <-ctx.Done():
			return
		}
	}

	wg.Wait()
}

func (c *Client) getAccrualSum(wg *sync.WaitGroup, order types.Order) {
	defer wg.Done()
	var accrualResp types.AccrualResponse
	url := fmt.Sprintf("%s%s%d", c.host, groupAPI, order.OrderID)

	resp, err := c.client.Get(url)
	switch {
	case err != nil:
		c.logger.WithError(err).Error("Get AccrualResponse client error")
		return
	case resp.StatusCode == 204:
		c.logger.Infof("No content for %d", order.OrderID)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&accrualResp)
	if err != nil {
		c.logger.WithError(err).Error("AccrualResponse decode client error")
		return
	}

	if order.Status != entities.StatusUnknown && order.Status != accrualResp.Status {
		c.chProcessingAccrual <- accrualResp
	}
}

func (c *Client) processAccrual(ctx context.Context) {
	for {
		select {
		case orderAccrual := <-c.chProcessingAccrual:
			err := c.storage.UpdateOrderAccrual(ctx, orderAccrual)
			if err != nil {
				c.logger.WithError(err).Error("processAccrual client storage error")
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

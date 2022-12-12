package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/iRootPro/gophermart/internal/store/sqlstore"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Accrual struct {
	Config *Config
	logger *logrus.Logger
	store  *sqlstore.Store
}

type ResponseAccrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewAccrual(config *Config, store *sqlstore.Store) *Accrual {
	return &Accrual{
		Config: config,
		logger: logrus.New(),
		store:  store,
	}
}

func (a *Accrual) configureLogger() error {
	level, err := logrus.ParseLevel(a.Config.LogLevel)
	if err != nil {
		return err
	}
	a.logger.SetLevel(level)
	return nil
}

func (a *Accrual) Run() {
	a.logger.Info("starting accrual service on address: ", a.Config.RunAddress)
	for {
		orders := a.store.Order().GetOrdersForUpgradeStatus()
		fmt.Printf("orders: %+v", orders)
		for _, order := range orders {
			response := a.GetOrder(order)
			err := a.store.Order().UpdateStatus(response.Order, response.Accrual, response.Status)
			if err != nil {
				a.logger.Error(err)
			}

			fmt.Println("updated")

		}
		time.Sleep(a.Config.PoolingTimeout)
	}
}

func (a *Accrual) GetOrder(orderNum string) ResponseAccrual {
	resp, err := http.Get(a.Config.RunAddress + "/api/orders/" + orderNum)
	if err != nil {
		a.logger.Error(err)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error(err)
	}

	defer resp.Body.Close()

	response := ResponseAccrual{}
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		a.logger.Error(err)
	}

	fmt.Printf("response: %+v", response)

	return response
}

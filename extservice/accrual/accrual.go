package accrual

import (
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
			a.GetOrder(order)
		}
		time.Sleep(a.Config.PoolingTimeout)
	}
}

func (a *Accrual) GetOrder(orderNum string) {
	resp, err := http.Get(a.Config.RunAddress + "/api/orders/" + orderNum)
	if err != nil {
		a.logger.Error(err)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Error(err)
	}

	fmt.Println(resBody)
}

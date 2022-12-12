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
		if len(orders) == 0 {
			continue
		}
		for _, order := range orders {
			response := a.GetOrder(order)
			err := a.store.Order().UpdateStatus(response.Order, response.Accrual, response.Status)
			if err != nil {
				a.logger.Error("update status", err)
			}

			if response.Status == "PROCESSED" {
				a.logger.Info("order: ", response.Order, " status: ", response.Status, " accrual: ", response.Accrual)
				userID, err := a.store.Order().FindUserIDByOrder(response.Order)
				if err != nil {
					a.logger.Error("find user_id by order number", err)
				}

				err = a.store.Balance().UpdateCurrentByUserID(userID, response.Accrual)
				if err != nil {
					a.logger.Error("update balance", err)
				}
				a.logger.Info("user_id: ", userID, " accrual: ", response.Accrual)
			}

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

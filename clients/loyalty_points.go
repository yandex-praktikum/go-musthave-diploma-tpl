package clients

import (
	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
	"net/http"
)

type LoyaltyPointsCalculationSystem struct {
}

func (c LoyaltyPointsCalculationSystem) DoGet(path string) (*http.Response, error) {
	return http.Get(config.DefaultClientDomain + "/api/orders/" + path)
}

var _ Client = &LoyaltyPointsCalculationSystem{}

func NewLoyaltyPointsCalculationSystem() LoyaltyPointsCalculationSystem {
	return LoyaltyPointsCalculationSystem{}
}

package db

import (
	"encoding/json"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"net/http"
)

func (d *DateBase) getAccrual(addressAccrual, order string) (*models.ResponseAccrual, error) {
	var accrual models.ResponseAccrual
	requestAccrual, err := http.Get(fmt.Sprintf("%s/api/orders/%s", addressAccrual, order))

	if err != nil {
		return nil, err
	}

	if err = json.NewDecoder(requestAccrual.Body).Decode(&accrual); err != nil {
		return nil, err
	}

	defer requestAccrual.Body.Close()

	return &accrual, nil
}

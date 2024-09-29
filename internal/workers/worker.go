package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"net/http"
	"strconv"
	"time"
)

type WorkerAccrual struct {
	storage *orders.Service
	log     *logger.Logger
}

func NewWorkerAccrual(storage *orders.Service, log *logger.Logger) *WorkerAccrual {
	return &WorkerAccrual{
		storage: storage,
		log:     log,
	}
}

func (w *WorkerAccrual) StartWorkerAccrual(ctx context.Context, addressAccrual string) {
	ticker := time.NewTicker(time.Second / 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go w.getAccrual(addressAccrual)
		case <-ctx.Done():
			return
		}
	}
}

func (w *WorkerAccrual) getAccrual(addressAccrual string) {
	query := "SELECT order_id, order_status FROM loyalty WHERE order_status IN ('REGISTERED', 'PROCESSING', 'NEW')" //берем только в статусе REGISTERED и PROCESSING и NEW
	rows, err := w.storage.Gets(query)
	if err != nil {
		w.log.Error("Error :", "gets rows in worker - ", customerrors.ErrNotData)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			w.log.Error("Error ", "closing row set:", err)
		}
	}()

	for rows.Next() {
		var (
			order       string
			orderStatus string
		)

		var accrual models.ResponseAccrual

		if err = rows.Scan(&order, &orderStatus); err != nil {
			w.log.Error("Error worker", "Error scanning rows in worker :", err)
			continue
		}

		req, err := http.Get(fmt.Sprintf("%s/api/orders/%s", addressAccrual, order))
		if err != nil {
			w.log.Error("Error worker", "Error making request in worker :", err)
			continue
		}
		defer req.Body.Close()

		if err = json.NewDecoder(req.Body).Decode(&accrual); err != nil {
			w.log.Error("Error worker", "Error decoding response in worker:", err)
			continue
		}

		if req.StatusCode == http.StatusTooManyRequests {
			timeSleep, err := strconv.Atoi(req.Header.Get("Retry-After"))
			if err != nil {
				time.Sleep(60 * time.Second)
			} else {
				time.Sleep(time.Duration(timeSleep) * time.Second)
			}
		}

		if orderStatus != accrual.Status {
			querySave := "UPDATE loyalty SET order_status = $1, bonus = $2 WHERE order_id = $3"
			err = w.storage.Save(querySave, accrual.Status, accrual.Accrual, order)
			if err != nil {
				w.log.Error("Error: ", "saving data in worker: ", err)
				continue
			}

			var loyaltyStatus string

			query := "SELECT order_status FROM loyalty WHERE order_id = $1"

			checkRow, err := w.storage.Get(query, accrual.Order)
			if err != nil {
				w.log.Error("Error worker", "Error check data in worker: ", err)
				continue
			}

			if err = checkRow.Scan(&loyaltyStatus); err != nil {
				w.log.Error("Error worker", "Error not found: ", err)
				continue
			}
		}
	}
}

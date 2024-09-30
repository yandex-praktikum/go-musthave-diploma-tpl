package workers

import (
	"context"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
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
			go w.storage.GetAccrual(addressAccrual)
		case <-ctx.Done():
			return
		}
	}
}

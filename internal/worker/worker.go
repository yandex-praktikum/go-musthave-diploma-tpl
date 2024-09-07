package worker

import (
	"context"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
)

var RequestAccrual = make(chan string, 10)

type WorkerAccrual struct {
	storage *service.Service
}

func NewWorkerAccrual(storage *service.Service) *WorkerAccrual {
	return &WorkerAccrual{
		storage: storage,
	}
}

func (w *WorkerAccrual) StartWorkerAccrual(ctx context.Context) {
	for {
		select {
		case req := <-RequestAccrual:
			go w.getAccrual(ctx, req)
		case <-ctx.Done():
			return
		}
	}
}

func (w *WorkerAccrual) getAccrual(ctx context.Context, req string) {
	
}

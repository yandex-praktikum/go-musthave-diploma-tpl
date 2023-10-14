package worker

import (
	"context"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/service/orders"
	"sync"
	"time"
)

type worker struct {
	workerCount int
	buffer      int
	wg          *sync.WaitGroup
	cancelFunc  context.CancelFunc
	order       orders.Order
}

type WorkerIface interface {
	Start(pctx context.Context)
	Stop()
	QueueTask(task string, workDuration time.Duration) error
}

func New(order orders.Order) WorkerIface {
	w := worker{
		wg:    new(sync.WaitGroup),
		order: order,
	}

	return &w
}

func (w *worker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.doWork(ctx)
		}
	}
}

func (w *worker) Start(pctx context.Context) {
	ctx, cancelFunc := context.WithCancel(pctx)
	w.cancelFunc = cancelFunc

	w.wg.Add(1)
	go w.spawnWorkers(ctx)
}

func (w *worker) Stop() {
	w.cancelFunc()
	w.wg.Wait()
}

func (w *worker) doWork(ctx context.Context) {
	w.order.AccrualOrderStatus(ctx)
	w.order.ChangeOrderStatus(ctx)
	// update task management data store indicating that the work is complete!
}

func sleepContext(ctx context.Context, sleep time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(sleep):
	}
}

func (w *worker) QueueTask(task string, workDuration time.Duration) error {
	return nil
}

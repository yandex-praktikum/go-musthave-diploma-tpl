package process

import (
	"context"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/clients"
	"github.com/k-morozov/go-musthave-diploma-tpl/core/entities"
	"github.com/rs/zerolog"
	"sync"
	"time"

	"github.com/k-morozov/go-musthave-diploma-tpl/components/config"
)

type TimerProcess struct {
	store          store.Store
	client         clients.Client
	log            zerolog.Logger
	readableTicket *time.Ticker
	done           chan struct{}
}

var _ Process = &TimerProcess{}

func NewTimerProcess(store store.Store, client clients.Client, log zerolog.Logger) Process {

	p := &TimerProcess{
		store:          store,
		client:         client,
		log:            log,
		readableTicket: time.NewTicker(config.DefaultReadableTicket * time.Second),
	}

	return p
}

func (p *TimerProcess) Run() {
	pullOrdersOutChan := p.runPullFromDB()
	fanOut := p.sendToClient(pullOrdersOutChan)
	fanIn := p.mergeResponses(fanOut...)
	p.applyDataToDB(fanIn)
}

// @TODO close?
func (p *TimerProcess) Stop() {
	p.done <- struct{}{}
}

func (p *TimerProcess) runPullFromDB() <-chan string {
	pullOrdersChan := make(chan string)

	go func() {
		defer p.readableTicket.Stop()
		defer close(pullOrdersChan)

		p.restorePulledNewOrders(pullOrdersChan)

		for {
			select {
			case _ = <-p.done:
				return
			case _ = <-p.readableTicket.C:
				p.pullNewOrders(pullOrdersChan)
			}
		}
	}()

	return pullOrdersChan
}

func (p *TimerProcess) restorePulledNewOrders(inChan chan<- string) {
	p.log.Debug().Msg("started restore processing")

	restoreOrders, err := p.store.RestoreNewOrders(context.TODO(), config.DefaultOwner)
	if err != nil {
		p.log.Err(err).Msg("error while locking new orders.")
		return
	}

	p.log.Debug().Msg("try send restored new orders")
	for _, order := range restoreOrders.Orders {
		select {
		case <-p.done:
			return
		case inChan <- order:

		}
	}

	p.log.Debug().Msg("start processing restored new orders")
}

func (p *TimerProcess) pullNewOrders(inChan chan<- string) {
	p.log.Debug().Msg("started processing")

	newOrders, err := p.store.GetNewOrders(context.TODO(), config.DefaultOwner, config.DefaultBatchSize)
	if err != nil {
		p.log.Err(err).Msg("error while locking new orders.")
		return
	}

	p.log.Debug().Msg("try send new orders")
	for _, order := range newOrders.Orders {
		select {
		case _ = <-p.done:
			return
		case inChan <- order:

		}
	}

	p.log.Debug().Msg("start processing new orders")
}

func (p *TimerProcess) sendToClient(pullOrdersOutChan <-chan string) []<-chan entities.GetOrderClientResponseBody {
	channels := make([]<-chan entities.GetOrderClientResponseBody, config.DefaultClientConnection)
	for i := 0; i < config.DefaultClientConnection; i++ {
		c := p.addSender(pullOrdersOutChan)
		channels[i] = c
	}
	return channels
}

func (p *TimerProcess) addSender(pullOrdersOutChan <-chan string) <-chan entities.GetOrderClientResponseBody {
	outCh := make(chan entities.GetOrderClientResponseBody)

	go func() {
		defer close(outCh)

		for orderID := range pullOrdersOutChan {
			p.log.Debug().Str("order_id", orderID).Msg("Start sending to client")
			r, err := p.client.DoGet(orderID)
			if err != nil {
				p.log.Err(err).Str("order_id", orderID).Msg("error from client")
				continue
			}
			res := entities.GetOrderClientResponseBody{}
			err = res.ParseFromResponse(r)
			if err != nil {
				p.log.Err(err).Str("order_id", orderID).Msg("failed parse response")
			}
			outCh <- res
			p.log.Debug().Any("result", res).Msg("send result to db")
		}

	}()

	return outCh
}

func (p *TimerProcess) mergeResponses(pullOrdersOutChan ...<-chan entities.GetOrderClientResponseBody) <-chan entities.GetOrderClientResponseBody {
	totalCh := make(chan entities.GetOrderClientResponseBody)
	var wg sync.WaitGroup

	for _, ch := range pullOrdersOutChan {
		localCh := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for result := range localCh {
				select {
				case _ = <-p.done:
					return
				case totalCh <- result:
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(totalCh)
	}()
	return totalCh
}

func (p *TimerProcess) applyDataToDB(inCh <-chan entities.GetOrderClientResponseBody) {
	go func() {
		for data := range inCh {
			err := p.store.ProcessedOrder(context.TODO(), models.NewOrderData(data.OrderID, data.Status, data.Accrual))
			if err != nil {
				// @TODO sad :(
			}
		}
	}()
}

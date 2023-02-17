package router

import (
	"fmt"
	"sync"
	"time"

	"GopherMart/internal/events"
)

func (s *serverMart) updateAccrual() {
	fmt.Println("=====updateAccrual===== ")
	var wg, wgTimer sync.WaitGroup
	for {
		orders, err := s.DB.ReadAllOrderAccrualNoComplite()
		if err != nil {
			//time.Sleep(1 * time.Second)
			continue
		}
		if len(orders) != 0 {
			for _, order := range orders {
				wg.Add(1)
				go s.worker(order.Order, order.Login, &wg, &wgTimer)
			}
		} else {
			//time.Sleep(1 * time.Second)
		}

		wg.Wait()
	}

}

func (s *serverMart) worker(order string, login string, wg, wgTimer *sync.WaitGroup) {
	wgTimer.Wait()
	accrual, sec, err := events.AccrualGet(s.Cfg.AccrualAddress, order)
	fmt.Println("=====1===== ", accrual, sec, err)
	for sec != 0 {
		wgTimer.Add(1)
		time.Sleep(time.Duration(sec) * time.Second)
		wgTimer.Done()
		accrual, sec, err = events.AccrualGet(s.Cfg.AccrualAddress, order)
	}
	fmt.Println("=====2===== ", accrual, sec, err)
	if err != nil {
		return
	}
	_ = s.DB.UpdateOrderAccrual(login, accrual)
	wg.Done()
}

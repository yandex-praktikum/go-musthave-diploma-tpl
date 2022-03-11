package repository

import (
	"sync"
)

type queue struct {
	queue []string
	mx    *sync.Mutex
}

func NewQueue() *queue {
	return &queue{
		queue: make([]string, 0, 10),
		mx:    &sync.Mutex{},
	}
}

//add to queue ===========================================================
func (q *queue) AddToQueue(order string) {
	q.mx.Lock()
	q.queue = append(q.queue, order)
	q.mx.Unlock()
}

//take first from queue ==================================================
func (q *queue) TakeFirst() string {
	var order string
	q.mx.Lock()
	if len(q.queue) != 0 {
		order = q.queue[0]
	}
	q.mx.Unlock()

	return order
}

//remove from queue =======================================================
func (q *queue) RemoveFromQueue() {
	q.mx.Lock()
	switch len(q.queue) {
	case 0:
		return
	case 1:
		q.queue = q.queue[:0]
	default:
		q.queue = q.queue[1:]
	}
	q.mx.Unlock()
}

package repository

//add to queue ===========================================================
func (r *Repository) AddToQueue(order string) {
	r.mx.Lock()
	r.queue = append(r.queue, order)
	r.mx.Unlock()
}

//take first from queue ==================================================
func (r *Repository) TakeFirst() string {
	var order string
	r.mx.Lock()
	if len(r.queue) != 0 {
		order = r.queue[0]
	}
	r.mx.Unlock()

	return order
}

//remove from queue =======================================================
func (r *Repository) RemoveFromQueue() {
	r.mx.Lock()
	switch len(r.queue) {
	case 0:
		return
	case 1:
		r.queue = r.queue[:0]
	default:
		r.queue = r.queue[1:]
	}
	r.mx.Unlock()
}

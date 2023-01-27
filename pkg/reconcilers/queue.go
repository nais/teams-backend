package reconcilers

import (
	"fmt"
	"sync"
)

const reconcilerQueueSize = 4096

type ReconcilerQueue interface {
	Add(input Input) error
	Close()
}

type queue struct {
	queue  chan Input
	closed bool
	lock   sync.Mutex
}

func NewReconcilerQueue() (ReconcilerQueue, <-chan Input) {
	ch := make(chan Input, reconcilerQueueSize)
	return &queue{
		queue:  ch,
		closed: false,
	}, ch
}

func (q *queue) Add(input Input) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.closed {
		return fmt.Errorf("team reconciler channel is closed")
	}
	q.queue <- input
	return nil
}

func (q *queue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true
	close(q.queue)
}

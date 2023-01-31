package teamsync

import (
	"fmt"
	"sync"

	"github.com/nais/console/pkg/reconcilers"
)

const reconcilerQueueSize = 4096

type Queue interface {
	Add(input reconcilers.Input) error
	Close()
}

type queue struct {
	queue  chan reconcilers.Input
	closed bool
	lock   sync.Mutex
}

func NewQueue() (Queue, <-chan reconcilers.Input) {
	ch := make(chan reconcilers.Input, reconcilerQueueSize)
	return &queue{
		queue:  ch,
		closed: false,
	}, ch
}

func (q *queue) Add(input reconcilers.Input) error {
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

package teamsync

import (
	"fmt"
	"sync"

	"github.com/nais/console/pkg/slug"
)

const reconcilerQueueSize = 4096

type Queue interface {
	Add(slug.Slug) error
	Close()
}

type queue struct {
	queue  chan slug.Slug
	closed bool
	lock   sync.Mutex
}

func NewQueue() (Queue, <-chan slug.Slug) {
	ch := make(chan slug.Slug, reconcilerQueueSize)
	return &queue{
		queue:  ch,
		closed: false,
	}, ch
}

func (q *queue) Add(slug slug.Slug) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.closed {
		return fmt.Errorf("team reconciler channel is closed")
	}

	q.queue <- slug
	return nil
}

func (q *queue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true
	close(q.queue)
}

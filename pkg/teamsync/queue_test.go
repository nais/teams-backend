package teamsync_test

import (
	"testing"

	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/teamsync"

	"github.com/stretchr/testify/assert"
)

func Test_Queue(t *testing.T) {
	q, ch := teamsync.NewQueue()
	slug := slug.Slug("slug")

	t.Run("add to queue", func(t *testing.T) {
		assert.Nil(t, q.Add(slug))
		assert.Len(t, ch, 1)
		assert.Equal(t, slug, <-ch)
		assert.Len(t, ch, 0)
	})

	t.Run("close channel", func(t *testing.T) {
		q.Close()
		assert.EqualError(t, q.Add(slug), "team reconciler channel is closed")
	})
}

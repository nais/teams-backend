package teamsync_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/teamsync"

	"github.com/stretchr/testify/assert"
)

func Test_Queue(t *testing.T) {
	q, ch := teamsync.NewQueue()
	input := teamsync.Input{
		TeamSlug:      slug.Slug("slug"),
		CorrelationID: uuid.New(),
	}

	t.Run("add to queue", func(t *testing.T) {
		assert.Nil(t, q.Add(input))
		assert.Len(t, ch, 1)
		assert.Equal(t, input, <-ch)
		assert.Len(t, ch, 0)
	})

	t.Run("close channel", func(t *testing.T) {
		q.Close()
		assert.EqualError(t, q.Add(input), "team reconciler channel is closed")
	})
}

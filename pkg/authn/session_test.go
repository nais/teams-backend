package authn_test

import (
	"github.com/nais/console/pkg/authn"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessionStore(t *testing.T) {
	t.Run("test set, get and delete", func(t *testing.T) {
		store := authn.NewStore()

		assert.Nil(t, store.Get("key"))

		session := &authn.Session{
			Key:     "key",
			Expires: time.Time{},
			Email:   "mail@example.com",
		}
		store.Create(session)
		assert.Equal(t, session, store.Get("key"))
		store.Destroy("key")
		assert.Nil(t, store.Get("key"))
	})
}

package slug_test

import (
	"bytes"
	"testing"

	"github.com/nais/teams-backend/pkg/slug"
	"github.com/stretchr/testify/assert"
)

func TestMarshalSlug(t *testing.T) {
	buf := new(bytes.Buffer)
	s := slug.Slug("some-slug")
	slug.MarshalSlug(&s).MarshalGQL(buf)
	assert.Equal(t, `"some-slug"`, buf.String())
}

func TestUnmarshalSlug(t *testing.T) {
	t.Run("invalid case", func(t *testing.T) {
		s, err := slug.UnmarshalSlug(123)
		assert.Nil(t, s)
		assert.EqualError(t, err, "slug must be a string")
	})

	t.Run("valid case", func(t *testing.T) {
		s, err := slug.UnmarshalSlug("slug")
		assert.NoError(t, err)
		assert.Equal(t, "slug", string(*s))
	})
}

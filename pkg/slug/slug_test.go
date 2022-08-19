package slug_test

import (
	"testing"

	"github.com/nais/console/pkg/slug"
	"github.com/stretchr/testify/assert"
)

func TestSlugValidation(t *testing.T) {
	assert.NoError(t, slug.Slug("slug").Validate())
	assert.NoError(t, slug.Slug("slug-with-dash").Validate())

	assert.Error(t, slug.Slug("slug-with-a-very-long-name").Validate())
	assert.Error(t, slug.Slug("slug-").Validate())
	assert.Error(t, slug.Slug("-slug").Validate())
	assert.Error(t, slug.Slug("slug-1").Validate())
}

package dbmodels_test

import (
	"testing"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/stretchr/testify/assert"
)

func TestSlugValidation(t *testing.T) {
	assert.NoError(t, dbmodels.Slug("slug").Validate())
	assert.NoError(t, dbmodels.Slug("slug-with-dash").Validate())

	assert.Error(t, dbmodels.Slug("slug-with-a-very-long-name").Validate())
	assert.Error(t, dbmodels.Slug("slug-").Validate())
	assert.Error(t, dbmodels.Slug("-slug").Validate())
	assert.Error(t, dbmodels.Slug("slug-1").Validate())
}

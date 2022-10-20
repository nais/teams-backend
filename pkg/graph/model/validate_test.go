package model_test

import (
	"testing"

	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/slug"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](value T) *T {
	return &value
}

func TestCreateTeamInput_Validate_Slug(t *testing.T) {
	tpl := model.CreateTeamInput{
		Slug:    nil,
		Name:    "valid name",
		Purpose: "valid purpose",
	}

	validSlugs := []string{
		"foo",
		"foo-bar",
		"f00b4r",
		"channel4",
		"some-long-string-less-than-31c",
	}

	invalidSlugs := []string{
		"a",
		"ab",
		"-foo",
		"foo-",
		"foo--bar",
		"4chan",
		"some-long-string-more-than-30-chars",
		"you-aint-got-the-æøå",
		"Uppercase",
		"rollback()",
	}

	for _, s := range validSlugs {
		tpl.Slug = ptr(slug.Slug(s))
		assert.NoError(t, tpl.Validate(), "Slug '%s' should pass validation, but didn't", tpl.Slug)
	}

	for _, s := range invalidSlugs {
		tpl.Slug = ptr(slug.Slug(s))
		assert.Error(t, tpl.Validate(), "Slug '%s' passed validation even if it should not", tpl.Slug)
	}
}

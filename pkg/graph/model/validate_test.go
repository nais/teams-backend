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

func TestCreateTeamInput_Validate_SlackAlertsChannel(t *testing.T) {
	tpl := model.CreateTeamInput{
		Slug:    ptr(slug.Slug("valid-slug")),
		Purpose: "valid purpose",
	}

	validChannels := []string{
		"#foo",
		"#foo-bar",
		"#æøå",
		"#aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	invalidChannels := []string{
		"foo", // missing hash
		"#a",  // too short
		"#aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", // too long
		"#foo bar", // space not allowed
		"#Foobar",  // upper case not allowed
	}

	for _, s := range validChannels {
		tpl.SlackAlertsChannel = s
		assert.NoError(t, tpl.Validate(), "Slack alerts channel %q should pass validation, but didn't", tpl.SlackAlertsChannel)
	}

	for _, s := range invalidChannels {
		tpl.SlackAlertsChannel = s
		assert.Error(t, tpl.Validate(), "Slack alerts channel %q passed validation even if it should not", tpl.SlackAlertsChannel)
	}
}

func TestCreateTeamInput_Validate_Slug(t *testing.T) {
	tpl := model.CreateTeamInput{
		Slug:               nil,
		Purpose:            "valid purpose",
		SlackAlertsChannel: "#channel",
	}

	validSlugs := []string{
		"foo",
		"foo-bar",
		"f00b4r",
		"channel4",
		"some-long-string-less-than-31c",
		"nais",
		"nais-system",
		"naisuratur",
		"naisan",
	}

	invalidSlugs := []string{
		"a",
		"ab",
		"-foo",
		"foo-",
		"foo--bar",
		"4chan",
		"team",
		"team-foo",
		"teamfoobar",
		"some-long-string-more-than-30-chars",
		"you-aint-got-the-æøå",
		"Uppercase",
		"rollback()",
	}

	for _, s := range validSlugs {
		tpl.Slug = ptr(slug.Slug(s))
		assert.NoError(t, tpl.Validate(), "Slug %q should pass validation, but didn't", tpl.Slug)
	}

	for _, s := range invalidSlugs {
		tpl.Slug = ptr(slug.Slug(s))
		assert.Error(t, tpl.Validate(), "Slug %q passed validation even if it should not", tpl.Slug)
	}
}

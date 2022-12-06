package model

import (
	"regexp"
	"strings"

	"github.com/nais/console/pkg/graph/apierror"
)

var (
	// Slightly modified from database schema because Golang doesn't like Perl-flavored regexes.
	teamSlugRegex = regexp.MustCompile("^[a-z](-?[a-z0-9]+)+$")

	// Rules can be found here: https://api.slack.com/methods/conversations.create#naming
	slackChannelNameRegex = regexp.MustCompile("^#[a-z0-9æøå_-]{2,80}$")
)

func ptr[T any](value T) *T {
	return &value
}

func (input CreateTeamInput) Validate() error {
	if input.Slug == nil || !teamSlugRegex.MatchString(string(*input.Slug)) || len(*input.Slug) < 3 || len(*input.Slug) > 30 {
		return apierror.ErrTeamSlug
	}

	if input.Purpose == "" {
		return apierror.ErrTeamPurpose
	}

	slug := input.Slug.String()

	if strings.HasPrefix(slug, "nais") {
		return apierror.ErrTeamPrefixReserved
	}

	if strings.HasPrefix(slug, "team") {
		return apierror.ErrTeamPrefixRedundant
	}

	if !slackChannelNameRegex.MatchString(input.SlackAlertsChannel) {
		return apierror.ErrTeamSlackAlertsChannel
	}

	return nil
}

func (input UpdateTeamInput) Validate() error {
	if input.Purpose != nil && *input.Purpose == "" {
		return apierror.ErrTeamPurpose
	}

	if input.SlackAlertsChannel != nil && !slackChannelNameRegex.MatchString(*input.SlackAlertsChannel) {
		return apierror.ErrTeamSlackAlertsChannel
	}

	return nil
}

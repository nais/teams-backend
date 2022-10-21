package model

import (
	"regexp"

	"github.com/nais/console/pkg/graph/apierror"
)

// Slightly modified from database schema because Golang doesn't like Perl-flavored regexes.
var teamSlugRegex = regexp.MustCompile("^[a-z](-?[a-z0-9]+)+$")

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

	return nil
}

func (input UpdateTeamInput) Validate() error {
	if input.Purpose != nil && *input.Purpose == "" {
		return apierror.ErrTeamPurpose
	}

	return nil
}

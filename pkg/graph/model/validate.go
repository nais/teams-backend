package model

import (
	"regexp"

	"github.com/nais/console/pkg/graph/apierror"
)

// Slightly modified from database schema because Golang doesn't like Perl-flavored regexes.
var teamSlugRegex = regexp.MustCompile("^[a-z](-?[a-z0-9]+){2,29}$")

func ptr[T any](value T) *T {
	return &value
}

func (input CreateTeamInput) Validate() error {
	if input.Slug == nil || !teamSlugRegex.MatchString(input.Slug.String()) {
		return apierror.ErrTeamSlug
	}

	if input.Name == "" {
		// FIXME: remove this field altogether
		return apierror.Errorf("You must specify a team name when creating a team.")
	}

	if input.Purpose == "" {
		return apierror.ErrTeamPurpose
	}

	return nil
}

func (input UpdateTeamInput) Validate() error {
	if input.Name != nil && *input.Name == "" {
		// FIXME: remove this field altogether
		return apierror.Errorf("You must specify a team name.")
	}

	if input.Purpose != nil && *input.Purpose == "" {
		return apierror.ErrTeamPurpose
	}

	return nil
}

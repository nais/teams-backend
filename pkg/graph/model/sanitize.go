package model

import (
	"strings"
)

func (input CreateTeamInput) Sanitize() CreateTeamInput {
	input.Purpose = strings.TrimSpace(input.Purpose)
	return input
}

func (input UpdateTeamInput) Sanitize() UpdateTeamInput {
	if input.Purpose != nil {
		input.Purpose = ptr(strings.TrimSpace(*input.Purpose))
	}
	return input
}

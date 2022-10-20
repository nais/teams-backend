package model

import (
	"strings"
)

func (input CreateTeamInput) Sanitize() CreateTeamInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Purpose = strings.TrimSpace(input.Purpose)
	return input
}

func (input UpdateTeamInput) Sanitize() UpdateTeamInput {
	if input.Name != nil {
		input.Name = ptr(strings.TrimSpace(*input.Name))
	}
	if input.Purpose != nil {
		input.Purpose = ptr(strings.TrimSpace(*input.Purpose))
	}
	return input
}

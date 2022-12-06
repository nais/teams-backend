package model

import (
	"strings"
)

func (input CreateTeamInput) Sanitize() CreateTeamInput {
	input.Purpose = strings.TrimSpace(input.Purpose)
	input.SlackAlertsChannel = strings.TrimSpace(input.SlackAlertsChannel)
	return input
}

func (input UpdateTeamInput) Sanitize() UpdateTeamInput {
	if input.Purpose != nil {
		input.Purpose = ptr(strings.TrimSpace(*input.Purpose))
	}

	if input.SlackAlertsChannel != nil {
		input.SlackAlertsChannel = ptr(strings.TrimSpace(*input.SlackAlertsChannel))
	}

	return input
}

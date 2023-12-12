package model

import (
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/slug"
)

type TeamsInternal struct{}

// Team member.
type TeamMember struct {
	TeamRole TeamRole
	TeamSlug slug.Slug
	UserID   uuid.UUID
}

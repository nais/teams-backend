package graph

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

func (r *mutationResolver) teamWithAssociations(teamID uuid.UUID) *dbmodels.Team {
	team := &dbmodels.Team{}
	r.db.Preload("Users").
		Preload("SystemState").
		Preload("TeamMetadata").
		First(team, "id = ?", teamID)
	return team
}

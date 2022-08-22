package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateTeam(ctx context.Context, teamID *uuid.UUID, input model.UpdateTeamInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, input model.RemoveUsersFromTeamInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) SynchronizeTeam(ctx context.Context, teamID *uuid.UUID) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddTeamMembers(ctx context.Context, input model.AddTeamMembersInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddTeamOwners(ctx context.Context, input model.AddTeamOwnersInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, input model.SetTeamMemberRoleInput) (*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Teams(ctx context.Context) ([]*db.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeams(ctx)
}

func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*db.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, *id)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeamByID(ctx, *id)
}

func (r *teamResolver) Purpose(ctx context.Context, obj *db.Team) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *teamResolver) Metadata(ctx context.Context, obj *db.Team) (map[string]interface{}, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]interface{})
	for k, v := range obj.Metadata {
		metadata[k] = v
	}

	return metadata, nil
}

func (r *teamResolver) AuditLogs(ctx context.Context, obj *db.Team) ([]*model.AuditLog, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *teamResolver) Members(ctx context.Context, obj *db.Team) ([]*model.TeamMember, error) {
	panic(fmt.Errorf("not implemented"))
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }

package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type TeamMetadata map[string]string

type Team struct {
	*sqlc.Team
}

func (d *database) GetTeamMetadata(ctx context.Context, teamID uuid.UUID) (TeamMetadata, error) {
	rows, err := d.querier.GetTeamMetadata(ctx, teamID)
	if err != nil {
		return nil, err
	}

	metadata := make(TeamMetadata)
	for _, row := range rows {
		metadata[row.Key] = row.Value.String
	}

	return metadata, nil
}

func (d *database) SetTeamMetadata(ctx context.Context, teamID uuid.UUID, metadata TeamMetadata) error {
	return d.querier.Transaction(ctx, func(querier Querier) error {
		for k, v := range metadata {
			err := querier.SetTeamMetadata(ctx, sqlc.SetTeamMetadataParams{
				TeamID: teamID,
				Key:    k,
				Value:  nullString(&v),
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *database) RemoveUserFromTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error {
	err := d.querier.RevokeTargetedRoleFromUser(ctx, sqlc.RevokeTargetedRoleFromUserParams{
		UserID:   userID,
		TargetID: nullUUID(&teamID),
		RoleName: sqlc.RoleNameTeammember,
	})
	if err != nil {
		return err
	}

	return d.querier.RevokeTargetedRoleFromUser(ctx, sqlc.RevokeTargetedRoleFromUserParams{
		UserID:   userID,
		TargetID: nullUUID(&teamID),
		RoleName: sqlc.RoleNameTeamowner,
	})
}

func (d *database) UpdateTeam(ctx context.Context, teamID uuid.UUID, name, purpose *string) (*Team, error) {
	if name != nil && *name == "" {
		return nil, fmt.Errorf("name can not be empty")
	}

	team, err := d.querier.UpdateTeam(ctx, sqlc.UpdateTeamParams{
		ID:      teamID,
		Name:    nullString(name),
		Purpose: nullString(purpose),
	})
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) AddTeam(ctx context.Context, name string, slug slug.Slug, purpose *string, ownerUserID uuid.UUID) (*Team, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	var team *sqlc.Team
	err = d.querier.Transaction(ctx, func(querier Querier) error {
		team, err = querier.CreateTeam(ctx, sqlc.CreateTeamParams{
			ID:      id,
			Name:    name,
			Slug:    slug,
			Purpose: nullString(purpose),
		})
		if err != nil {
			return err
		}

		err = querier.AssignTargetedRoleToUser(ctx, sqlc.AssignTargetedRoleToUserParams{
			UserID:   ownerUserID,
			TargetID: nullUUID(&team.ID),
			RoleName: sqlc.RoleNameTeamowner,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error) {
	team, err := d.querier.GetTeamBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	team, err := d.querier.GetTeamByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) GetTeams(ctx context.Context) ([]*Team, error) {
	teams, err := d.querier.GetTeams(ctx)
	if err != nil {
		return nil, err
	}

	collection := make([]*Team, 0)
	for _, team := range teams {
		collection = append(collection, &Team{Team: team})
	}

	return collection, nil
}

func (d *database) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	rows, err := d.querier.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	teams := make([]*Team, 0)
	for _, team := range rows {
		teams = append(teams, &Team{Team: team})
	}

	return teams, nil
}

func (d *database) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error) {
	mems, err := d.querier.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	members := make([]*User, 0)
	for _, m := range mems {
		members = append(members, &User{User: m})
	}

	return members, nil
}

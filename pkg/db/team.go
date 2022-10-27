package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) GetTeamMetadata(ctx context.Context, teamID uuid.UUID) ([]*TeamMetadata, error) {
	rows, err := d.querier.GetTeamMetadata(ctx, teamID)
	if err != nil {
		return nil, err
	}

	metadata := make([]*TeamMetadata, 0, len(rows))
	for _, row := range rows {
		var value *string
		if row.Value.Valid {
			value = &row.Value.String
		}
		metadata = append(metadata, &TeamMetadata{
			Key:   row.Key,
			Value: value,
		})
	}

	return metadata, nil
}

func (d *database) SetTeamMetadata(ctx context.Context, teamID uuid.UUID, metadata []TeamMetadata) error {
	return d.querier.Transaction(ctx, func(ctx context.Context, querier Querier) error {
		for _, entry := range metadata {
			err := querier.SetTeamMetadata(ctx, sqlc.SetTeamMetadataParams{
				TeamID: teamID,
				Key:    entry.Key,
				Value:  nullString(entry.Value),
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

func (d *database) UpdateTeam(ctx context.Context, teamID uuid.UUID, purpose *string) (*Team, error) {
	team, err := d.querier.UpdateTeam(ctx, sqlc.UpdateTeamParams{
		ID:      teamID,
		Purpose: nullString(purpose),
	})
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) CreateTeam(ctx context.Context, slug slug.Slug, purpose string) (*Team, error) {
	team, err := d.querier.CreateTeam(ctx, sqlc.CreateTeamParams{
		Slug:    slug,
		Purpose: purpose,
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
	rows, err := d.querier.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	members := make([]*User, 0)
	for _, row := range rows {
		members = append(members, &User{User: row})
	}

	return members, nil
}

func (d *database) DisableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	team, err := d.querier.DisableTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) EnableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	team, err := d.querier.EnableTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type Team struct {
	*sqlc.Team
	Metadata map[string]string
	Members  []*User
}

func (d *database) AddUserToTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error {
	return d.querier.AddUserToTeam(ctx, sqlc.AddUserToTeamParams{
		UserID: userID,
		TeamID: teamID,
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

func (d *database) AddTeam(ctx context.Context, name, slug string, purpose *string, userID uuid.UUID) (*Team, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	querier := d.querier.WithTx(tx)
	team, err := querier.CreateTeam(ctx, sqlc.CreateTeamParams{
		ID:      id,
		Name:    name,
		Slug:    slug,
		Purpose: nullString(purpose),
	})
	if err != nil {
		return nil, err
	}

	err = querier.AddUserToTeam(ctx, sqlc.AddUserToTeamParams{
		UserID: userID,
		TeamID: team.ID,
	})
	if err != nil {
		return nil, err
	}

	err = querier.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
		UserID:   userID,
		RoleName: sqlc.RoleNameTeamowner,
		TargetID: nullUUID(&team.ID),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) GetTeamBySlug(ctx context.Context, slug string) (*Team, error) {
	team, err := d.querier.GetTeamBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return d.getTeam(ctx, &Team{Team: team})
}

func (d *database) GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	team, err := d.querier.GetTeamByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return d.getTeam(ctx, &Team{Team: team})
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

func (d *database) getTeam(ctx context.Context, team *Team) (*Team, error) {
	metadata, err := d.querier.GetTeamMetadata(ctx, team.ID)
	if err != nil {
		return nil, err
	}

	if team.Metadata == nil {
		team.Metadata = make(map[string]string)
	}
	for _, row := range metadata {
		team.Metadata[row.Key] = row.Value.String
	}

	return team, nil
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

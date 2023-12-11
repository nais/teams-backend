package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
)

func (d *database) RemoveUserFromTeam(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) error {
	return d.querier.RemoveUserFromTeam(ctx, sqlc.RemoveUserFromTeamParams{
		UserID:         userID,
		TargetTeamSlug: &teamSlug,
	})
}

func (d *database) UpdateTeam(ctx context.Context, teamSlug slug.Slug, purpose, slackChannel *string) (*Team, error) {
	team, err := d.querier.UpdateTeam(ctx, sqlc.UpdateTeamParams{
		Slug:         teamSlug,
		Purpose:      purpose,
		SlackChannel: slackChannel,
	})
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) CreateTeam(ctx context.Context, slug slug.Slug, purpose, slackChannel string) (*Team, error) {
	team, err := d.querier.CreateTeam(ctx, sqlc.CreateTeamParams{
		Slug:         slug,
		Purpose:      purpose,
		SlackChannel: slackChannel,
	})
	if err != nil {
		return nil, err
	}

	return &Team{Team: team}, nil
}

func (d *database) GetActiveTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error) {
	team, err := d.querier.GetActiveTeamBySlug(ctx, slug)
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

func (d *database) GetTeams(ctx context.Context, offset, limit *int) ([]*Team, error) {
	var teams []*sqlc.Team
	var err error
	if limit != nil {
		if offset == nil {
			o := 0
			offset = &o
		}
		teams, err = d.querier.GetTeamsPaginated(ctx, sqlc.GetTeamsPaginatedParams{
			Limit:  int32(*limit),
			Offset: int32(*offset),
		})
	} else {
		teams, err = d.querier.GetTeams(ctx)
	}
	if err != nil {
		return nil, err
	}

	collection := make([]*Team, 0)
	for _, team := range teams {
		collection = append(collection, &Team{Team: team})
	}

	return collection, nil
}

func (d *database) GetActiveTeams(ctx context.Context) ([]*Team, error) {
	teams, err := d.querier.GetActiveTeams(ctx)
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

func (d *database) GetTeamMembers(ctx context.Context, teamSlug slug.Slug, offset, limit *int) ([]*User, error) {
	var rows []*sqlc.User
	var err error
	if limit != nil {
		if offset == nil {
			o := 0
			offset = &o
		}
		rows, err = d.querier.GetTeamMembersPaginated(ctx, sqlc.GetTeamMembersPaginatedParams{
			TargetTeamSlug: &teamSlug,
			Limit:          int32(*limit),
			Offset:         int32(*offset),
		})
	} else {
		rows, err = d.querier.GetTeamMembers(ctx, &teamSlug)
	}
	if err != nil {
		return nil, err
	}

	members := make([]*User, 0)
	for _, row := range rows {
		members = append(members, &User{User: row})
	}

	return members, nil
}

func (d *database) GetTeamMember(ctx context.Context, teamSlug slug.Slug, userID uuid.UUID) (*User, error) {
	user, err := d.querier.GetTeamMember(ctx, sqlc.GetTeamMemberParams{
		TargetTeamSlug: &teamSlug,
		ID:             userID,
	})
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetTeamMembersForReconciler(ctx context.Context, teamSlug slug.Slug, reconcilerName sqlc.ReconcilerName) ([]*User, error) {
	rows, err := d.querier.GetTeamMembersForReconciler(ctx, sqlc.GetTeamMembersForReconcilerParams{
		TargetTeamSlug: &teamSlug,
		ReconcilerName: reconcilerName,
	})
	if err != nil {
		return nil, err
	}

	members := make([]*User, 0)
	for _, row := range rows {
		members = append(members, &User{User: row})
	}

	return members, nil
}

func (d *database) SetLastSuccessfulSyncForTeam(ctx context.Context, teamSlug slug.Slug) error {
	return d.querier.SetLastSuccessfulSyncForTeam(ctx, teamSlug)
}

func (d *database) GetSlackAlertsChannels(ctx context.Context, teamSlug slug.Slug) (map[string]string, error) {
	channels := make(map[string]string)
	rows, err := d.querier.GetSlackAlertsChannels(ctx, teamSlug)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		channels[row.Environment] = row.ChannelName
	}

	return channels, nil
}

func (d *database) SetSlackAlertsChannel(ctx context.Context, teamSlug slug.Slug, environment, channelName string) error {
	return d.querier.SetSlackAlertsChannel(ctx, sqlc.SetSlackAlertsChannelParams{
		TeamSlug:    teamSlug,
		Environment: environment,
		ChannelName: channelName,
	})
}

func (d *database) RemoveSlackAlertsChannel(ctx context.Context, teamSlug slug.Slug, environment string) error {
	return d.querier.RemoveSlackAlertsChannel(ctx, sqlc.RemoveSlackAlertsChannelParams{
		TeamSlug:    teamSlug,
		Environment: environment,
	})
}

func (d *database) CreateTeamDeleteKey(ctx context.Context, teamSlug slug.Slug, userID uuid.UUID) (*TeamDeleteKey, error) {
	deleteKey, err := d.querier.CreateTeamDeleteKey(ctx, sqlc.CreateTeamDeleteKeyParams{
		TeamSlug:  teamSlug,
		CreatedBy: userID,
	})
	if err != nil {
		return nil, err
	}
	return &TeamDeleteKey{TeamDeleteKey: deleteKey}, nil
}

func (d *database) GetTeamDeleteKey(ctx context.Context, key uuid.UUID) (*TeamDeleteKey, error) {
	deleteKey, err := d.querier.GetTeamDeleteKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return &TeamDeleteKey{TeamDeleteKey: deleteKey}, nil
}

func (d *database) ConfirmTeamDeleteKey(ctx context.Context, key uuid.UUID) error {
	return d.querier.ConfirmTeamDeleteKey(ctx, key)
}

func (d *database) DeleteTeam(ctx context.Context, teamSlug slug.Slug) error {
	return d.querier.DeleteTeam(ctx, teamSlug)
}

func (d *database) GetTeamMemberOptOuts(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) ([]*sqlc.GetTeamMemberOptOutsRow, error) {
	return d.querier.GetTeamMemberOptOuts(ctx, sqlc.GetTeamMemberOptOutsParams{
		UserID:   userID,
		TeamSlug: teamSlug,
	})
}

func (d *database) GetTeamsWithPermissionInGitHubRepo(ctx context.Context, repoName, permission string) ([]*Team, error) {
	var state pgtype.JSONB
	err := state.Set(map[string]interface{}{
		"repositories": []map[string]interface{}{
			{
				"name": repoName,
				"permissions": []map[string]interface{}{
					{
						"name":    permission,
						"granted": true,
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	rows, err := d.querier.GetTeamsWithPermissionInGitHubRepo(ctx, state)
	if err != nil {
		return nil, err
	}

	teams := make([]*Team, 0)
	for _, row := range rows {
		teams = append(teams, &Team{row})
	}

	return teams, nil
}

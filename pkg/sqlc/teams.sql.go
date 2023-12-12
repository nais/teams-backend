// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: teams.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/slug"
)

const confirmTeamDeleteKey = `-- name: ConfirmTeamDeleteKey :exec
UPDATE team_delete_keys
SET confirmed_at = NOW()
WHERE key = $1
`

func (q *Queries) ConfirmTeamDeleteKey(ctx context.Context, key uuid.UUID) error {
	_, err := q.db.Exec(ctx, confirmTeamDeleteKey, key)
	return err
}

const createTeam = `-- name: CreateTeam :one
INSERT INTO teams (slug, purpose, slack_channel)
VALUES ($1, $2, $3)
RETURNING slug, purpose, last_successful_sync, slack_channel
`

type CreateTeamParams struct {
	Slug         slug.Slug
	Purpose      string
	SlackChannel string
}

func (q *Queries) CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error) {
	row := q.db.QueryRow(ctx, createTeam, arg.Slug, arg.Purpose, arg.SlackChannel)
	var i Team
	err := row.Scan(
		&i.Slug,
		&i.Purpose,
		&i.LastSuccessfulSync,
		&i.SlackChannel,
	)
	return &i, err
}

const createTeamDeleteKey = `-- name: CreateTeamDeleteKey :one
INSERT INTO team_delete_keys (team_slug, created_by)
VALUES($1, $2)
RETURNING key, team_slug, created_at, created_by, confirmed_at
`

type CreateTeamDeleteKeyParams struct {
	TeamSlug  slug.Slug
	CreatedBy uuid.UUID
}

func (q *Queries) CreateTeamDeleteKey(ctx context.Context, arg CreateTeamDeleteKeyParams) (*TeamDeleteKey, error) {
	row := q.db.QueryRow(ctx, createTeamDeleteKey, arg.TeamSlug, arg.CreatedBy)
	var i TeamDeleteKey
	err := row.Scan(
		&i.Key,
		&i.TeamSlug,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.ConfirmedAt,
	)
	return &i, err
}

const deleteTeam = `-- name: DeleteTeam :exec
DELETE FROM teams
WHERE slug = $1
`

func (q *Queries) DeleteTeam(ctx context.Context, argSlug slug.Slug) error {
	_, err := q.db.Exec(ctx, deleteTeam, argSlug)
	return err
}

const getActiveTeamBySlug = `-- name: GetActiveTeamBySlug :one
SELECT teams.slug, teams.purpose, teams.last_successful_sync, teams.slack_channel FROM teams
WHERE
    teams.slug = $1
    AND NOT EXISTS (
        SELECT team_delete_keys.team_slug
        FROM team_delete_keys
        WHERE
            team_delete_keys.team_slug = $1
            AND team_delete_keys.confirmed_at IS NOT NULL
    )
`

func (q *Queries) GetActiveTeamBySlug(ctx context.Context, argSlug slug.Slug) (*Team, error) {
	row := q.db.QueryRow(ctx, getActiveTeamBySlug, argSlug)
	var i Team
	err := row.Scan(
		&i.Slug,
		&i.Purpose,
		&i.LastSuccessfulSync,
		&i.SlackChannel,
	)
	return &i, err
}

const getActiveTeams = `-- name: GetActiveTeams :many
SELECT teams.slug, teams.purpose, teams.last_successful_sync, teams.slack_channel FROM teams
WHERE NOT EXISTS (
    SELECT team_delete_keys.team_slug
    FROM team_delete_keys
    WHERE
        team_delete_keys.team_slug = teams.slug
        AND team_delete_keys.confirmed_at IS NOT NULL
)
ORDER BY teams.slug ASC
`

func (q *Queries) GetActiveTeams(ctx context.Context) ([]*Team, error) {
	rows, err := q.db.Query(ctx, getActiveTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Team
	for rows.Next() {
		var i Team
		if err := rows.Scan(
			&i.Slug,
			&i.Purpose,
			&i.LastSuccessfulSync,
			&i.SlackChannel,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSlackAlertsChannels = `-- name: GetSlackAlertsChannels :many
SELECT team_slug, environment, channel_name FROM slack_alerts_channels
WHERE team_slug = $1
ORDER BY environment ASC
`

func (q *Queries) GetSlackAlertsChannels(ctx context.Context, teamSlug slug.Slug) ([]*SlackAlertsChannel, error) {
	rows, err := q.db.Query(ctx, getSlackAlertsChannels, teamSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SlackAlertsChannel
	for rows.Next() {
		var i SlackAlertsChannel
		if err := rows.Scan(&i.TeamSlug, &i.Environment, &i.ChannelName); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeamBySlug = `-- name: GetTeamBySlug :one
SELECT teams.slug, teams.purpose, teams.last_successful_sync, teams.slack_channel FROM teams
WHERE teams.slug = $1
`

func (q *Queries) GetTeamBySlug(ctx context.Context, argSlug slug.Slug) (*Team, error) {
	row := q.db.QueryRow(ctx, getTeamBySlug, argSlug)
	var i Team
	err := row.Scan(
		&i.Slug,
		&i.Purpose,
		&i.LastSuccessfulSync,
		&i.SlackChannel,
	)
	return &i, err
}

const getTeamDeleteKey = `-- name: GetTeamDeleteKey :one
SELECT key, team_slug, created_at, created_by, confirmed_at FROM team_delete_keys
WHERE key = $1
`

func (q *Queries) GetTeamDeleteKey(ctx context.Context, key uuid.UUID) (*TeamDeleteKey, error) {
	row := q.db.QueryRow(ctx, getTeamDeleteKey, key)
	var i TeamDeleteKey
	err := row.Scan(
		&i.Key,
		&i.TeamSlug,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.ConfirmedAt,
	)
	return &i, err
}

const getTeamMember = `-- name: GetTeamMember :one
SELECT users.id, users.email, users.name, users.external_id FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1 AND users.id = $2
ORDER BY users.name ASC
`

type GetTeamMemberParams struct {
	TargetTeamSlug *slug.Slug
	ID             uuid.UUID
}

func (q *Queries) GetTeamMember(ctx context.Context, arg GetTeamMemberParams) (*User, error) {
	row := q.db.QueryRow(ctx, getTeamMember, arg.TargetTeamSlug, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

const getTeamMemberOptOuts = `-- name: GetTeamMemberOptOuts :many
SELECT
    reconcilers.name,
    NOT EXISTS(
        SELECT reconciler_name FROM reconciler_opt_outs
        WHERE user_id = $1 AND team_slug = $2 AND reconciler_name = reconcilers.name
    ) AS enabled
FROM reconcilers
WHERE reconcilers.enabled = true
ORDER BY reconcilers.name ASC
`

type GetTeamMemberOptOutsParams struct {
	UserID   uuid.UUID
	TeamSlug slug.Slug
}

type GetTeamMemberOptOutsRow struct {
	Name    ReconcilerName
	Enabled bool
}

func (q *Queries) GetTeamMemberOptOuts(ctx context.Context, arg GetTeamMemberOptOutsParams) ([]*GetTeamMemberOptOutsRow, error) {
	rows, err := q.db.Query(ctx, getTeamMemberOptOuts, arg.UserID, arg.TeamSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetTeamMemberOptOutsRow
	for rows.Next() {
		var i GetTeamMemberOptOutsRow
		if err := rows.Scan(&i.Name, &i.Enabled); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeamMembers = `-- name: GetTeamMembers :many
SELECT users.id, users.email, users.name, users.external_id FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1
ORDER BY users.name ASC
`

func (q *Queries) GetTeamMembers(ctx context.Context, targetTeamSlug *slug.Slug) ([]*User, error) {
	rows, err := q.db.Query(ctx, getTeamMembers, targetTeamSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.Name,
			&i.ExternalID,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeamMembersForReconciler = `-- name: GetTeamMembersForReconciler :many
SELECT users.id, users.email, users.name, users.external_id FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE
    user_roles.target_team_slug = $1
    AND NOT EXISTS (
        SELECT roo.user_id
        FROM reconciler_opt_outs AS roo
        WHERE
            roo.team_slug = $1
            AND roo.reconciler_name = $2
            AND roo.user_id = users.id
    )
ORDER BY users.name ASC
`

type GetTeamMembersForReconcilerParams struct {
	TargetTeamSlug *slug.Slug
	ReconcilerName ReconcilerName
}

func (q *Queries) GetTeamMembersForReconciler(ctx context.Context, arg GetTeamMembersForReconcilerParams) ([]*User, error) {
	rows, err := q.db.Query(ctx, getTeamMembersForReconciler, arg.TargetTeamSlug, arg.ReconcilerName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.Name,
			&i.ExternalID,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeamMembersPaginated = `-- name: GetTeamMembersPaginated :many
SELECT users.id, users.email, users.name, users.external_id FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1
ORDER BY users.name ASC LIMIT $2 OFFSET $3
`

type GetTeamMembersPaginatedParams struct {
	TargetTeamSlug *slug.Slug
	Limit          int32
	Offset         int32
}

func (q *Queries) GetTeamMembersPaginated(ctx context.Context, arg GetTeamMembersPaginatedParams) ([]*User, error) {
	rows, err := q.db.Query(ctx, getTeamMembersPaginated, arg.TargetTeamSlug, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.Name,
			&i.ExternalID,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeams = `-- name: GetTeams :many
SELECT teams.slug, teams.purpose, teams.last_successful_sync, teams.slack_channel FROM teams
ORDER BY teams.slug ASC
`

func (q *Queries) GetTeams(ctx context.Context) ([]*Team, error) {
	rows, err := q.db.Query(ctx, getTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Team
	for rows.Next() {
		var i Team
		if err := rows.Scan(
			&i.Slug,
			&i.Purpose,
			&i.LastSuccessfulSync,
			&i.SlackChannel,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTeamsCount = `-- name: GetTeamsCount :one
SELECT COUNT(*) as total FROM teams
`

func (q *Queries) GetTeamsCount(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getTeamsCount)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const getTeamsPaginated = `-- name: GetTeamsPaginated :many
SELECT teams.slug, teams.purpose, teams.last_successful_sync, teams.slack_channel FROM teams
ORDER BY teams.slug ASC LIMIT $1 OFFSET $2
`

type GetTeamsPaginatedParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) GetTeamsPaginated(ctx context.Context, arg GetTeamsPaginatedParams) ([]*Team, error) {
	rows, err := q.db.Query(ctx, getTeamsPaginated, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Team
	for rows.Next() {
		var i Team
		if err := rows.Scan(
			&i.Slug,
			&i.Purpose,
			&i.LastSuccessfulSync,
			&i.SlackChannel,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeSlackAlertsChannel = `-- name: RemoveSlackAlertsChannel :exec
DELETE FROM slack_alerts_channels
WHERE team_slug = $1 AND environment = $2
`

type RemoveSlackAlertsChannelParams struct {
	TeamSlug    slug.Slug
	Environment string
}

func (q *Queries) RemoveSlackAlertsChannel(ctx context.Context, arg RemoveSlackAlertsChannelParams) error {
	_, err := q.db.Exec(ctx, removeSlackAlertsChannel, arg.TeamSlug, arg.Environment)
	return err
}

const removeUserFromTeam = `-- name: RemoveUserFromTeam :exec
DELETE FROM user_roles
WHERE user_id = $1 AND target_team_slug = $2
`

type RemoveUserFromTeamParams struct {
	UserID         uuid.UUID
	TargetTeamSlug *slug.Slug
}

func (q *Queries) RemoveUserFromTeam(ctx context.Context, arg RemoveUserFromTeamParams) error {
	_, err := q.db.Exec(ctx, removeUserFromTeam, arg.UserID, arg.TargetTeamSlug)
	return err
}

const setLastSuccessfulSyncForTeam = `-- name: SetLastSuccessfulSyncForTeam :exec
UPDATE teams SET last_successful_sync = NOW()
WHERE slug = $1
`

func (q *Queries) SetLastSuccessfulSyncForTeam(ctx context.Context, argSlug slug.Slug) error {
	_, err := q.db.Exec(ctx, setLastSuccessfulSyncForTeam, argSlug)
	return err
}

const setSlackAlertsChannel = `-- name: SetSlackAlertsChannel :exec
INSERT INTO slack_alerts_channels (team_slug, environment, channel_name)
VALUES ($1, $2, $3)
ON CONFLICT (team_slug, environment) DO
    UPDATE SET channel_name = $3
`

type SetSlackAlertsChannelParams struct {
	TeamSlug    slug.Slug
	Environment string
	ChannelName string
}

func (q *Queries) SetSlackAlertsChannel(ctx context.Context, arg SetSlackAlertsChannelParams) error {
	_, err := q.db.Exec(ctx, setSlackAlertsChannel, arg.TeamSlug, arg.Environment, arg.ChannelName)
	return err
}

const updateTeam = `-- name: UpdateTeam :one
UPDATE teams
SET purpose = COALESCE($1, purpose),
    slack_channel = COALESCE($2, slack_channel)
WHERE slug = $3
RETURNING slug, purpose, last_successful_sync, slack_channel
`

type UpdateTeamParams struct {
	Purpose      *string
	SlackChannel *string
	Slug         slug.Slug
}

func (q *Queries) UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error) {
	row := q.db.QueryRow(ctx, updateTeam, arg.Purpose, arg.SlackChannel, arg.Slug)
	var i Team
	err := row.Scan(
		&i.Slug,
		&i.Purpose,
		&i.LastSuccessfulSync,
		&i.SlackChannel,
	)
	return &i, err
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: teams.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
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

func (q *Queries) DeleteTeam(ctx context.Context, slug slug.Slug) error {
	_, err := q.db.Exec(ctx, deleteTeam, slug)
	return err
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

func (q *Queries) GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error) {
	row := q.db.QueryRow(ctx, getTeamBySlug, slug)
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

const getTeamMetadata = `-- name: GetTeamMetadata :many
SELECT key, value, team_slug FROM team_metadata
WHERE team_slug = $1
ORDER BY key ASC
`

func (q *Queries) GetTeamMetadata(ctx context.Context, teamSlug slug.Slug) ([]*TeamMetadatum, error) {
	rows, err := q.db.Query(ctx, getTeamMetadata, teamSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TeamMetadatum
	for rows.Next() {
		var i TeamMetadatum
		if err := rows.Scan(&i.Key, &i.Value, &i.TeamSlug); err != nil {
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

func (q *Queries) SetLastSuccessfulSyncForTeam(ctx context.Context, slug slug.Slug) error {
	_, err := q.db.Exec(ctx, setLastSuccessfulSyncForTeam, slug)
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

const setTeamMetadata = `-- name: SetTeamMetadata :exec
INSERT INTO team_metadata (team_slug, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (team_slug, key) DO
    UPDATE SET value = $3
`

type SetTeamMetadataParams struct {
	TeamSlug slug.Slug
	Key      string
	Value    sql.NullString
}

func (q *Queries) SetTeamMetadata(ctx context.Context, arg SetTeamMetadataParams) error {
	_, err := q.db.Exec(ctx, setTeamMetadata, arg.TeamSlug, arg.Key, arg.Value)
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
	Purpose      sql.NullString
	SlackChannel sql.NullString
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

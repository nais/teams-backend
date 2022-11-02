// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: teams.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

const createTeam = `-- name: CreateTeam :one
INSERT INTO teams (slug, purpose)
VALUES ($1, $2)
RETURNING id, slug, purpose, enabled
`

type CreateTeamParams struct {
	Slug    slug.Slug
	Purpose string
}

func (q *Queries) CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error) {
	row := q.db.QueryRow(ctx, createTeam, arg.Slug, arg.Purpose)
	var i Team
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Purpose,
		&i.Enabled,
	)
	return &i, err
}

const disableTeam = `-- name: DisableTeam :one
UPDATE teams
SET enabled = false
WHERE id = $1
RETURNING id, slug, purpose, enabled
`

func (q *Queries) DisableTeam(ctx context.Context, id uuid.UUID) (*Team, error) {
	row := q.db.QueryRow(ctx, disableTeam, id)
	var i Team
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Purpose,
		&i.Enabled,
	)
	return &i, err
}

const enableTeam = `-- name: EnableTeam :one
UPDATE teams
SET enabled = true
WHERE id = $1
RETURNING id, slug, purpose, enabled
`

func (q *Queries) EnableTeam(ctx context.Context, id uuid.UUID) (*Team, error) {
	row := q.db.QueryRow(ctx, enableTeam, id)
	var i Team
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Purpose,
		&i.Enabled,
	)
	return &i, err
}

const getTeamBySlug = `-- name: GetTeamBySlug :one
SELECT id, slug, purpose, enabled FROM teams
WHERE slug = $1
`

func (q *Queries) GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error) {
	row := q.db.QueryRow(ctx, getTeamBySlug, slug)
	var i Team
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Purpose,
		&i.Enabled,
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
SELECT id, slug, purpose, enabled FROM teams
ORDER BY slug ASC
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
			&i.ID,
			&i.Slug,
			&i.Purpose,
			&i.Enabled,
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
SET purpose = COALESCE($1, purpose)
WHERE id = $2
RETURNING id, slug, purpose, enabled
`

type UpdateTeamParams struct {
	Purpose sql.NullString
	ID      uuid.UUID
}

func (q *Queries) UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error) {
	row := q.db.QueryRow(ctx, updateTeam, arg.Purpose, arg.ID)
	var i Team
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Purpose,
		&i.Enabled,
	)
	return &i, err
}

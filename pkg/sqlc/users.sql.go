// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: users.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, email, external_id)
VALUES ($1, LOWER($2), $3)
RETURNING id, email, name, external_id
`

type CreateUserParams struct {
	Name       string
	Email      string
	ExternalID string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Name, arg.Email, arg.ExternalID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, name, external_id FROM users
WHERE email = LOWER($1)
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

const getUserByExternalID = `-- name: GetUserByExternalID :one
SELECT id, email, name, external_id FROM users
WHERE external_id = $1
`

func (q *Queries) GetUserByExternalID(ctx context.Context, externalID string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByExternalID, externalID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, name, external_id FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

const getUserTeams = `-- name: GetUserTeams :many
SELECT teams.slug, teams.purpose, teams.enabled, teams.last_successful_sync, teams.slack_channel FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.user_id = $1
ORDER BY teams.slug ASC
`

func (q *Queries) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	rows, err := q.db.Query(ctx, getUserTeams, userID)
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
			&i.Enabled,
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

const getUsers = `-- name: GetUsers :many
SELECT id, email, name, external_id FROM users
ORDER BY name ASC
`

func (q *Queries) GetUsers(ctx context.Context) ([]*User, error) {
	rows, err := q.db.Query(ctx, getUsers)
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

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET name = $1, email = LOWER($4), external_id = $2
WHERE id = $3
RETURNING id, email, name, external_id
`

type UpdateUserParams struct {
	Name       string
	ExternalID string
	ID         uuid.UUID
	Email      string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.Name,
		arg.ExternalID,
		arg.ID,
		arg.Email,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.ExternalID,
	)
	return &i, err
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: users.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const addGlobalUserRole = `-- name: AddGlobalUserRole :exec
INSERT INTO user_roles (user_id, role_name) VALUES ($1, $2) ON CONFLICT DO NOTHING
`

type AddGlobalUserRoleParams struct {
	UserID   uuid.UUID
	RoleName RoleName
}

func (q *Queries) AddGlobalUserRole(ctx context.Context, arg AddGlobalUserRoleParams) error {
	_, err := q.db.Exec(ctx, addGlobalUserRole, arg.UserID, arg.RoleName)
	return err
}

const addTargetedUserRole = `-- name: AddTargetedUserRole :exec
INSERT INTO user_roles (user_id, role_name, target_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
`

type AddTargetedUserRoleParams struct {
	UserID   uuid.UUID
	RoleName RoleName
	TargetID uuid.NullUUID
}

func (q *Queries) AddTargetedUserRole(ctx context.Context, arg AddTargetedUserRoleParams) error {
	_, err := q.db.Exec(ctx, addTargetedUserRole, arg.UserID, arg.RoleName, arg.TargetID)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
RETURNING id, email, name
`

type CreateUserParams struct {
	ID    uuid.UUID
	Name  string
	Email string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.ID, arg.Name, arg.Email)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name)
	return &i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getUserByApiKey = `-- name: GetUserByApiKey :one
SELECT users.id, users.email, users.name FROM api_keys
JOIN users ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1 LIMIT 1
`

func (q *Queries) GetUserByApiKey(ctx context.Context, apiKey string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByApiKey, apiKey)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name)
	return &i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, name FROM users
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name)
	return &i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, name FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name)
	return &i, err
}

const getUserTeams = `-- name: GetUserTeams :many
SELECT teams.id, teams.slug, teams.name, teams.purpose FROM user_roles JOIN teams ON teams.id = user_roles.target_id WHERE user_roles.user_id = $1 ORDER BY teams.name ASC
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
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Purpose,
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
SELECT id, email, name FROM users
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
		if err := rows.Scan(&i.ID, &i.Email, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsersByEmail = `-- name: GetUsersByEmail :many
SELECT id, email, name FROM users
WHERE email LIKE $1 LIMIT 1
`

func (q *Queries) GetUsersByEmail(ctx context.Context, email string) ([]*User, error) {
	rows, err := q.db.Query(ctx, getUsersByEmail, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.ID, &i.Email, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeAllUserRoles = `-- name: RemoveAllUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1
`

func (q *Queries) RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.Exec(ctx, removeAllUserRoles, userID)
	return err
}

const removeGlobalUserRole = `-- name: RemoveGlobalUserRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND target_id IS NULL AND role_name = $2
`

type RemoveGlobalUserRoleParams struct {
	UserID   uuid.UUID
	RoleName RoleName
}

func (q *Queries) RemoveGlobalUserRole(ctx context.Context, arg RemoveGlobalUserRoleParams) error {
	_, err := q.db.Exec(ctx, removeGlobalUserRole, arg.UserID, arg.RoleName)
	return err
}

const removeTargetedUserRole = `-- name: RemoveTargetedUserRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND target_id = $2 AND role_name = $3
`

type RemoveTargetedUserRoleParams struct {
	UserID   uuid.UUID
	TargetID uuid.NullUUID
	RoleName RoleName
}

func (q *Queries) RemoveTargetedUserRole(ctx context.Context, arg RemoveTargetedUserRoleParams) error {
	_, err := q.db.Exec(ctx, removeTargetedUserRole, arg.UserID, arg.TargetID, arg.RoleName)
	return err
}

const setUserName = `-- name: SetUserName :one
UPDATE users SET name = $1 WHERE id = $2
RETURNING id, email, name
`

type SetUserNameParams struct {
	Name string
	ID   uuid.UUID
}

func (q *Queries) SetUserName(ctx context.Context, arg SetUserNameParams) (*User, error) {
	row := q.db.QueryRow(ctx, setUserName, arg.Name, arg.ID)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name)
	return &i, err
}

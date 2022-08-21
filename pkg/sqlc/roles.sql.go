// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: roles.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const getRoleAuthorizations = `-- name: GetRoleAuthorizations :many
SELECT authz_name
FROM role_authz
WHERE role_name = $1
ORDER BY authz_name ASC
`

func (q *Queries) GetRoleAuthorizations(ctx context.Context, roleName RoleName) ([]AuthzName, error) {
	rows, err := q.db.Query(ctx, getRoleAuthorizations, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthzName
	for rows.Next() {
		var authz_name AuthzName
		if err := rows.Scan(&authz_name); err != nil {
			return nil, err
		}
		items = append(items, authz_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoleNames = `-- name: GetRoleNames :many
SELECT unnest(enum_range(NULL::role_name))::role_name
`

func (q *Queries) GetRoleNames(ctx context.Context) ([]RoleName, error) {
	rows, err := q.db.Query(ctx, getRoleNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RoleName
	for rows.Next() {
		var column_1 RoleName
		if err := rows.Scan(&column_1); err != nil {
			return nil, err
		}
		items = append(items, column_1)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserRole = `-- name: GetUserRole :one
SELECT id, role_name, user_id, target_id FROM user_roles
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserRole(ctx context.Context, id uuid.UUID) (*UserRole, error) {
	row := q.db.QueryRow(ctx, getUserRole, id)
	var i UserRole
	err := row.Scan(
		&i.ID,
		&i.RoleName,
		&i.UserID,
		&i.TargetID,
	)
	return &i, err
}

const getUserRoles = `-- name: GetUserRoles :many
SELECT id, role_name, user_id, target_id FROM user_roles
WHERE user_id = $1
`

func (q *Queries) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error) {
	rows, err := q.db.Query(ctx, getUserRoles, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*UserRole
	for rows.Next() {
		var i UserRole
		if err := rows.Scan(
			&i.ID,
			&i.RoleName,
			&i.UserID,
			&i.TargetID,
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

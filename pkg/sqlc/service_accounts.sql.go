// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: service_accounts.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createServiceAccount = `-- name: CreateServiceAccount :one
INSERT INTO service_accounts (name)
VALUES ($1)
RETURNING id, name
`

func (q *Queries) CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	row := q.db.QueryRow(ctx, createServiceAccount, name)
	var i ServiceAccount
	err := row.Scan(&i.ID, &i.Name)
	return &i, err
}

const deleteServiceAccount = `-- name: DeleteServiceAccount :exec
DELETE FROM service_accounts
WHERE id = $1
`

func (q *Queries) DeleteServiceAccount(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteServiceAccount, id)
	return err
}

const getServiceAccountByApiKey = `-- name: GetServiceAccountByApiKey :one
SELECT service_accounts.id, service_accounts.name FROM api_keys
JOIN service_accounts ON service_accounts.id = api_keys.service_account_id
WHERE api_keys.api_key = $1
`

func (q *Queries) GetServiceAccountByApiKey(ctx context.Context, apiKey string) (*ServiceAccount, error) {
	row := q.db.QueryRow(ctx, getServiceAccountByApiKey, apiKey)
	var i ServiceAccount
	err := row.Scan(&i.ID, &i.Name)
	return &i, err
}

const getServiceAccountByName = `-- name: GetServiceAccountByName :one
SELECT id, name FROM service_accounts
WHERE name = $1
`

func (q *Queries) GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error) {
	row := q.db.QueryRow(ctx, getServiceAccountByName, name)
	var i ServiceAccount
	err := row.Scan(&i.ID, &i.Name)
	return &i, err
}

const getServiceAccountRoles = `-- name: GetServiceAccountRoles :many
SELECT id, role_name, service_account_id, target_id FROM service_account_roles
WHERE service_account_id = $1
`

func (q *Queries) GetServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) ([]*ServiceAccountRole, error) {
	rows, err := q.db.Query(ctx, getServiceAccountRoles, serviceAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ServiceAccountRole
	for rows.Next() {
		var i ServiceAccountRole
		if err := rows.Scan(
			&i.ID,
			&i.RoleName,
			&i.ServiceAccountID,
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

const getServiceAccounts = `-- name: GetServiceAccounts :many
SELECT id, name FROM service_accounts
ORDER BY name ASC
`

func (q *Queries) GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error) {
	rows, err := q.db.Query(ctx, getServiceAccounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ServiceAccount
	for rows.Next() {
		var i ServiceAccount
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
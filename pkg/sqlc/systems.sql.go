// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: systems.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createSystem = `-- name: CreateSystem :one
INSERT INTO systems (id, name) VALUES ($1, $2)
RETURNING id, created_at, created_by_id, updated_by_id, updated_at, name
`

type CreateSystemParams struct {
	ID   uuid.UUID
	Name string
}

func (q *Queries) CreateSystem(ctx context.Context, arg CreateSystemParams) (*System, error) {
	row := q.queryRow(ctx, q.createSystemStmt, createSystem, arg.ID, arg.Name)
	var i System
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.UpdatedAt,
		&i.Name,
	)
	return &i, err
}

const getSystem = `-- name: GetSystem :one
SELECT id, created_at, created_by_id, updated_by_id, updated_at, name FROM systems
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetSystem(ctx context.Context, id uuid.UUID) (*System, error) {
	row := q.queryRow(ctx, q.getSystemStmt, getSystem, id)
	var i System
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.UpdatedAt,
		&i.Name,
	)
	return &i, err
}

const getSystemByName = `-- name: GetSystemByName :one
SELECT id, created_at, created_by_id, updated_by_id, updated_at, name FROM systems
WHERE name = $1 LIMIT 1
`

func (q *Queries) GetSystemByName(ctx context.Context, name string) (*System, error) {
	row := q.queryRow(ctx, q.getSystemByNameStmt, getSystemByName, name)
	var i System
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.UpdatedAt,
		&i.Name,
	)
	return &i, err
}

const getSystems = `-- name: GetSystems :many
SELECT id, created_at, created_by_id, updated_by_id, updated_at, name FROM systems
ORDER BY name ASC
`

func (q *Queries) GetSystems(ctx context.Context) ([]*System, error) {
	rows, err := q.query(ctx, q.getSystemsStmt, getSystems)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*System
	for rows.Next() {
		var i System
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.CreatedByID,
			&i.UpdatedByID,
			&i.UpdatedAt,
			&i.Name,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

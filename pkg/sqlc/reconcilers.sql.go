// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: reconcilers.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

const addReconcilerOptOut = `-- name: AddReconcilerOptOut :exec
INSERT INTO reconciler_opt_outs (team_slug, user_id, reconciler_name)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
`

type AddReconcilerOptOutParams struct {
	TeamSlug       slug.Slug
	UserID         uuid.UUID
	ReconcilerName ReconcilerName
}

func (q *Queries) AddReconcilerOptOut(ctx context.Context, arg AddReconcilerOptOutParams) error {
	_, err := q.db.Exec(ctx, addReconcilerOptOut, arg.TeamSlug, arg.UserID, arg.ReconcilerName)
	return err
}

const configureReconciler = `-- name: ConfigureReconciler :exec
UPDATE reconciler_config
SET value = $3::TEXT
WHERE reconciler = $1 AND key = $2
`

type ConfigureReconcilerParams struct {
	Reconciler ReconcilerName
	Key        ReconcilerConfigKey
	Value      string
}

func (q *Queries) ConfigureReconciler(ctx context.Context, arg ConfigureReconcilerParams) error {
	_, err := q.db.Exec(ctx, configureReconciler, arg.Reconciler, arg.Key, arg.Value)
	return err
}

const dangerousGetReconcilerConfigValues = `-- name: DangerousGetReconcilerConfigValues :many
SELECT key, value::TEXT
FROM reconciler_config
WHERE reconciler = $1
`

type DangerousGetReconcilerConfigValuesRow struct {
	Key   ReconcilerConfigKey
	Value string
}

func (q *Queries) DangerousGetReconcilerConfigValues(ctx context.Context, reconciler ReconcilerName) ([]*DangerousGetReconcilerConfigValuesRow, error) {
	rows, err := q.db.Query(ctx, dangerousGetReconcilerConfigValues, reconciler)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*DangerousGetReconcilerConfigValuesRow
	for rows.Next() {
		var i DangerousGetReconcilerConfigValuesRow
		if err := rows.Scan(&i.Key, &i.Value); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const disableReconciler = `-- name: DisableReconciler :one
UPDATE reconcilers
SET enabled = false
WHERE name = $1
RETURNING name, display_name, description, enabled, run_order
`

func (q *Queries) DisableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error) {
	row := q.db.QueryRow(ctx, disableReconciler, name)
	var i Reconciler
	err := row.Scan(
		&i.Name,
		&i.DisplayName,
		&i.Description,
		&i.Enabled,
		&i.RunOrder,
	)
	return &i, err
}

const enableReconciler = `-- name: EnableReconciler :one
UPDATE reconcilers
SET enabled = true
WHERE name = $1
RETURNING name, display_name, description, enabled, run_order
`

func (q *Queries) EnableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error) {
	row := q.db.QueryRow(ctx, enableReconciler, name)
	var i Reconciler
	err := row.Scan(
		&i.Name,
		&i.DisplayName,
		&i.Description,
		&i.Enabled,
		&i.RunOrder,
	)
	return &i, err
}

const getEnabledReconcilers = `-- name: GetEnabledReconcilers :many
SELECT name, display_name, description, enabled, run_order FROM reconcilers
WHERE enabled = true
ORDER BY run_order ASC
`

func (q *Queries) GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error) {
	rows, err := q.db.Query(ctx, getEnabledReconcilers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Reconciler
	for rows.Next() {
		var i Reconciler
		if err := rows.Scan(
			&i.Name,
			&i.DisplayName,
			&i.Description,
			&i.Enabled,
			&i.RunOrder,
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

const getReconciler = `-- name: GetReconciler :one
SELECT name, display_name, description, enabled, run_order FROM reconcilers
WHERE name = $1
`

func (q *Queries) GetReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error) {
	row := q.db.QueryRow(ctx, getReconciler, name)
	var i Reconciler
	err := row.Scan(
		&i.Name,
		&i.DisplayName,
		&i.Description,
		&i.Enabled,
		&i.RunOrder,
	)
	return &i, err
}

const getReconcilerConfig = `-- name: GetReconcilerConfig :many
SELECT
    rc.reconciler,
    rc.key,
    rc.display_name,
    rc.description,
    (rc.value IS NOT NULL)::BOOL AS configured,
    rc2.value,
    rc.secret
FROM reconciler_config rc
LEFT JOIN reconciler_config rc2 ON rc2.reconciler = rc.reconciler AND rc2.key = rc.key AND rc2.secret = false
WHERE rc.reconciler = $1
ORDER BY rc.display_name ASC
`

type GetReconcilerConfigRow struct {
	Reconciler  ReconcilerName
	Key         ReconcilerConfigKey
	DisplayName string
	Description string
	Configured  bool
	Value       *string
	Secret      bool
}

func (q *Queries) GetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) ([]*GetReconcilerConfigRow, error) {
	rows, err := q.db.Query(ctx, getReconcilerConfig, reconciler)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetReconcilerConfigRow
	for rows.Next() {
		var i GetReconcilerConfigRow
		if err := rows.Scan(
			&i.Reconciler,
			&i.Key,
			&i.DisplayName,
			&i.Description,
			&i.Configured,
			&i.Value,
			&i.Secret,
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

const getReconcilers = `-- name: GetReconcilers :many
SELECT name, display_name, description, enabled, run_order FROM reconcilers
ORDER BY run_order ASC
`

func (q *Queries) GetReconcilers(ctx context.Context) ([]*Reconciler, error) {
	rows, err := q.db.Query(ctx, getReconcilers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Reconciler
	for rows.Next() {
		var i Reconciler
		if err := rows.Scan(
			&i.Name,
			&i.DisplayName,
			&i.Description,
			&i.Enabled,
			&i.RunOrder,
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

const removeReconcilerOptOut = `-- name: RemoveReconcilerOptOut :exec
DELETE FROM reconciler_opt_outs
WHERE team_slug = $1 AND user_id = $2 AND reconciler_name = $3
`

type RemoveReconcilerOptOutParams struct {
	TeamSlug       slug.Slug
	UserID         uuid.UUID
	ReconcilerName ReconcilerName
}

func (q *Queries) RemoveReconcilerOptOut(ctx context.Context, arg RemoveReconcilerOptOutParams) error {
	_, err := q.db.Exec(ctx, removeReconcilerOptOut, arg.TeamSlug, arg.UserID, arg.ReconcilerName)
	return err
}

const resetReconcilerConfig = `-- name: ResetReconcilerConfig :exec
UPDATE reconciler_config
SET value = NULL
WHERE reconciler = $1
`

func (q *Queries) ResetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) error {
	_, err := q.db.Exec(ctx, resetReconcilerConfig, reconciler)
	return err
}

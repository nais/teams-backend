// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: reconciler_states.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/nais/console/pkg/slug"
)

const getReconcilerStateForTeam = `-- name: GetReconcilerStateForTeam :one
SELECT reconciler, state, team_slug FROM reconciler_states
WHERE reconciler = $1 AND team_slug = $2
`

type GetReconcilerStateForTeamParams struct {
	Reconciler ReconcilerName
	TeamSlug   slug.Slug
}

func (q *Queries) GetReconcilerStateForTeam(ctx context.Context, arg GetReconcilerStateForTeamParams) (*ReconcilerState, error) {
	row := q.db.QueryRow(ctx, getReconcilerStateForTeam, arg.Reconciler, arg.TeamSlug)
	var i ReconcilerState
	err := row.Scan(&i.Reconciler, &i.State, &i.TeamSlug)
	return &i, err
}

const removeReconcilerStateForTeam = `-- name: RemoveReconcilerStateForTeam :exec
DELETE FROM reconciler_states
WHERE reconciler = $1 AND team_slug = $2
`

type RemoveReconcilerStateForTeamParams struct {
	Reconciler ReconcilerName
	TeamSlug   slug.Slug
}

func (q *Queries) RemoveReconcilerStateForTeam(ctx context.Context, arg RemoveReconcilerStateForTeamParams) error {
	_, err := q.db.Exec(ctx, removeReconcilerStateForTeam, arg.Reconciler, arg.TeamSlug)
	return err
}

const setReconcilerStateForTeam = `-- name: SetReconcilerStateForTeam :exec
INSERT INTO reconciler_states (reconciler, team_slug, state)
VALUES($1, $2, $3)
ON CONFLICT (reconciler, team_slug) DO
    UPDATE SET state = $3
`

type SetReconcilerStateForTeamParams struct {
	Reconciler ReconcilerName
	TeamSlug   slug.Slug
	State      pgtype.JSONB
}

func (q *Queries) SetReconcilerStateForTeam(ctx context.Context, arg SetReconcilerStateForTeamParams) error {
	_, err := q.db.Exec(ctx, setReconcilerStateForTeam, arg.Reconciler, arg.TeamSlug, arg.State)
	return err
}

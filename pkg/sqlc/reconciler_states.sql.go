// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: reconciler_states.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const getReconcilerStateForTeam = `-- name: GetReconcilerStateForTeam :one
SELECT system_name, team_id, state FROM reconciler_states
WHERE system_name = $1 AND team_id = $2
`

type GetReconcilerStateForTeamParams struct {
	SystemName SystemName
	TeamID     uuid.UUID
}

func (q *Queries) GetReconcilerStateForTeam(ctx context.Context, arg GetReconcilerStateForTeamParams) (*ReconcilerState, error) {
	row := q.db.QueryRow(ctx, getReconcilerStateForTeam, arg.SystemName, arg.TeamID)
	var i ReconcilerState
	err := row.Scan(&i.SystemName, &i.TeamID, &i.State)
	return &i, err
}

const setReconcilerStateForTeam = `-- name: SetReconcilerStateForTeam :exec
INSERT INTO reconciler_states (system_name, team_id, state)
VALUES($1, $2, $3)
ON CONFLICT (system_name, team_id) DO
    UPDATE SET state = $3
`

type SetReconcilerStateForTeamParams struct {
	SystemName SystemName
	TeamID     uuid.UUID
	State      pgtype.JSONB
}

func (q *Queries) SetReconcilerStateForTeam(ctx context.Context, arg SetReconcilerStateForTeamParams) error {
	_, err := q.db.Exec(ctx, setReconcilerStateForTeam, arg.SystemName, arg.TeamID, arg.State)
	return err
}

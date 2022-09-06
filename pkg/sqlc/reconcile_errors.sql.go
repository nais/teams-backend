// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: reconcile_errors.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const clearTeamReconcileErrorForSystem = `-- name: ClearTeamReconcileErrorForSystem :exec
DELETE FROM reconcile_errors WHERE team_id = $1 AND system_name = $2
`

type ClearTeamReconcileErrorForSystemParams struct {
	TeamID     uuid.UUID
	SystemName SystemName
}

func (q *Queries) ClearTeamReconcileErrorForSystem(ctx context.Context, arg ClearTeamReconcileErrorForSystemParams) error {
	_, err := q.db.Exec(ctx, clearTeamReconcileErrorForSystem, arg.TeamID, arg.SystemName)
	return err
}

const getTeamReconcileErrors = `-- name: GetTeamReconcileErrors :many
SELECT id, correlation_id, team_id, system_name, created_at, error_message FROM reconcile_errors WHERE team_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetTeamReconcileErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcileError, error) {
	rows, err := q.db.Query(ctx, getTeamReconcileErrors, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ReconcileError
	for rows.Next() {
		var i ReconcileError
		if err := rows.Scan(
			&i.ID,
			&i.CorrelationID,
			&i.TeamID,
			&i.SystemName,
			&i.CreatedAt,
			&i.ErrorMessage,
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

const setTeamReconcileErrorForSystem = `-- name: SetTeamReconcileErrorForSystem :exec
INSERT INTO reconcile_errors (correlation_id, team_id, system_name, error_message)
VALUES ($1, $2, $3, $4)
ON CONFLICT(team_id, system_name) DO
    UPDATE SET correlation_id = $1, created_at = NOW(), error_message = $4
`

type SetTeamReconcileErrorForSystemParams struct {
	CorrelationID uuid.UUID
	TeamID        uuid.UUID
	SystemName    SystemName
	ErrorMessage  string
}

func (q *Queries) SetTeamReconcileErrorForSystem(ctx context.Context, arg SetTeamReconcileErrorForSystemParams) error {
	_, err := q.db.Exec(ctx, setTeamReconcileErrorForSystem,
		arg.CorrelationID,
		arg.TeamID,
		arg.SystemName,
		arg.ErrorMessage,
	)
	return err
}

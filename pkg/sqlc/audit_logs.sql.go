// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: audit_logs.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createAuditLog = `-- name: CreateAuditLog :one
INSERT INTO audit_logs (id, correlation_id, actor_email, system_name, target_user_email, target_team_slug, action, message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, correlation_id, actor_email, system_name, target_user_email, target_team_slug, action, message
`

type CreateAuditLogParams struct {
	ID              uuid.UUID
	CorrelationID   uuid.UUID
	ActorEmail      sql.NullString
	SystemName      NullSystemName
	TargetUserEmail sql.NullString
	TargetTeamSlug  sql.NullString
	Action          AuditAction
	Message         string
}

func (q *Queries) CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) (*AuditLog, error) {
	row := q.db.QueryRow(ctx, createAuditLog,
		arg.ID,
		arg.CorrelationID,
		arg.ActorEmail,
		arg.SystemName,
		arg.TargetUserEmail,
		arg.TargetTeamSlug,
		arg.Action,
		arg.Message,
	)
	var i AuditLog
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.CorrelationID,
		&i.ActorEmail,
		&i.SystemName,
		&i.TargetUserEmail,
		&i.TargetTeamSlug,
		&i.Action,
		&i.Message,
	)
	return &i, err
}
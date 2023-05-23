// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: audit_logs.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createAuditLog = `-- name: CreateAuditLog :exec
INSERT INTO audit_logs (correlation_id, actor, system_name, target_type, target_identifier, action, message)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type CreateAuditLogParams struct {
	CorrelationID    uuid.UUID
	Actor            sql.NullString
	SystemName       SystemName
	TargetType       AuditLogsTargetType
	TargetIdentifier string
	Action           AuditAction
	Message          string
}

func (q *Queries) CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error {
	_, err := q.db.Exec(ctx, createAuditLog,
		arg.CorrelationID,
		arg.Actor,
		arg.SystemName,
		arg.TargetType,
		arg.TargetIdentifier,
		arg.Action,
		arg.Message,
	)
	return err
}

const getAuditLogsForCorrelationID = `-- name: GetAuditLogsForCorrelationID :many
SELECT id, created_at, correlation_id, system_name, actor, action, message, target_type, target_identifier FROM audit_logs
WHERE correlation_id = $1
ORDER BY created_at DESC
`

func (q *Queries) GetAuditLogsForCorrelationID(ctx context.Context, correlationID uuid.UUID) ([]*AuditLog, error) {
	rows, err := q.db.Query(ctx, getAuditLogsForCorrelationID, correlationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AuditLog
	for rows.Next() {
		var i AuditLog
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.CorrelationID,
			&i.SystemName,
			&i.Actor,
			&i.Action,
			&i.Message,
			&i.TargetType,
			&i.TargetIdentifier,
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

const getAuditLogsForReconciler = `-- name: GetAuditLogsForReconciler :many
SELECT id, created_at, correlation_id, system_name, actor, action, message, target_type, target_identifier FROM audit_logs
WHERE target_type = 'reconciler' AND target_identifier = $1
ORDER BY created_at DESC
LIMIT 100
`

func (q *Queries) GetAuditLogsForReconciler(ctx context.Context, targetIdentifier string) ([]*AuditLog, error) {
	rows, err := q.db.Query(ctx, getAuditLogsForReconciler, targetIdentifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AuditLog
	for rows.Next() {
		var i AuditLog
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.CorrelationID,
			&i.SystemName,
			&i.Actor,
			&i.Action,
			&i.Message,
			&i.TargetType,
			&i.TargetIdentifier,
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

const getAuditLogsForTeam = `-- name: GetAuditLogsForTeam :many
SELECT id, created_at, correlation_id, system_name, actor, action, message, target_type, target_identifier FROM audit_logs
WHERE target_type = 'team' AND target_identifier = $1
ORDER BY created_at DESC
LIMIT 100
`

func (q *Queries) GetAuditLogsForTeam(ctx context.Context, targetIdentifier string) ([]*AuditLog, error) {
	rows, err := q.db.Query(ctx, getAuditLogsForTeam, targetIdentifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AuditLog
	for rows.Next() {
		var i AuditLog
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.CorrelationID,
			&i.SystemName,
			&i.Actor,
			&i.Action,
			&i.Message,
			&i.TargetType,
			&i.TargetIdentifier,
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

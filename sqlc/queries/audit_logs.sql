-- name: CreateAuditLog :exec
INSERT INTO audit_logs (correlation_id, actor_email, system_name, target_user_email, target_team_slug, action, message)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetAuditLogsForTeam :many
SELECT * FROM audit_logs WHERE target_team_slug = $1 ORDER BY created_at DESC LIMIT 100;
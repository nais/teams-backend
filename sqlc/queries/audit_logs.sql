-- name: CreateAuditLog :one
INSERT INTO audit_logs (id, correlation_id, actor_email, system_name, target_user_email, target_team_slug, action, message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

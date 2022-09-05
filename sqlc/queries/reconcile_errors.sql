-- name: PurgeReconcileError :exec
DELETE FROM reconcile_errors WHERE team_id = $1 AND system_name = $2;

-- name: AddReconcileError :exec
INSERT INTO reconcile_errors (correlation_id, team_id, system_name, error_message)
VALUES ($1, $2, $3, $4)
ON CONFLICT(team_id, system_name) DO
    UPDATE SET correlation_id = $1, created_at = NOW(), error_message = $4;

-- name: GetReconcileErrorsForTeam :many
SELECT * FROM reconcile_errors WHERE team_id = $1 ORDER BY created_at DESC;
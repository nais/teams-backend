-- name: GetReconcilers :many
SELECT * FROM reconcilers
ORDER BY run_order ASC;

-- name: GetEnabledReconcilers :many
SELECT * FROM reconcilers
WHERE enabled = true
ORDER BY run_order ASC;

-- name: GetReconciler :one
SELECT * FROM reconcilers
WHERE name = $1;

-- name: EnableReconciler :one
UPDATE reconcilers
SET enabled = true
WHERE name = $1 AND enabled = false
RETURNING *;

-- name: DisableReconciler :one
UPDATE reconcilers
SET enabled = false
WHERE name = $1 AND enabled = true
RETURNING *;

-- name: ResetReconcilerConfig :exec
UPDATE reconciler_config
SET value = NULL
WHERE reconciler = $1;

-- name: ConfigureReconciler :exec
UPDATE reconciler_config
SET value = sqlc.arg(value)::TEXT
WHERE reconciler = $1 AND key = $2;

-- name: GetReconcilerConfig :many
SELECT reconciler, key, display_name, description, (value IS NOT NULL)::BOOL AS configured, (CASE WHEN secret = false THEN value ELSE NULL END) AS value, secret
FROM reconciler_config
WHERE reconciler = $1;

-- name: DangerousGetReconcilerConfigValues :many
SELECT key, value::TEXT
FROM reconciler_config
WHERE reconciler = $1;
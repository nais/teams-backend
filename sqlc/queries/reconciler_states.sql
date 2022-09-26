-- name: GetReconcilerStateForTeam :one
SELECT * FROM reconciler_states
WHERE reconciler = $1 AND team_id = $2;

-- name: SetReconcilerStateForTeam :exec
INSERT INTO reconciler_states (reconciler, team_id, state)
VALUES($1, $2, $3)
ON CONFLICT (reconciler, team_id) DO
    UPDATE SET state = $3;
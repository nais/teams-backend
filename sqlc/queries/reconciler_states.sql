-- name: GetReconcilerStateForTeam :one
SELECT * FROM reconciler_states
WHERE system_name = $1 AND team_id = $2;

-- name: SetReconcilerStateForTeam :exec
INSERT INTO reconciler_states (system_name, team_id, state)
VALUES($1, $2, $3)
ON CONFLICT (system_name, team_id) DO
    UPDATE SET state = $3;
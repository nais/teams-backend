-- name: GetTeamSystemState :one
SELECT * FROM system_states WHERE system_name = $1 AND team_id = $2 LIMIT 1;

-- name: SetTeamSystemState :exec
INSERT INTO system_states (system_name, team_id, state) VALUES($1, $2, $3) ON CONFLICT (system_name, team_id) DO UPDATE SET state = $3;
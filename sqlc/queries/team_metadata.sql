-- name: GetTeamMetadata :one
SELECT * FROM team_metadata WHERE team_id = $1 LIMIT 1;
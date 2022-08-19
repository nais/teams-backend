-- name: GetTeamMetadata :many
SELECT * FROM team_metadata WHERE team_id = $1;
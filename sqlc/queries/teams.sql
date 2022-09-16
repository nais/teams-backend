-- name: CreateTeam :one
INSERT INTO teams (name, slug, purpose)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeams :many
SELECT * FROM teams
ORDER BY name ASC;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1;

-- name: GetTeamBySlug :one
SELECT * FROM teams
WHERE slug = $1;

-- name: GetTeamMembers :many
SELECT users.* FROM user_roles
JOIN teams ON teams.id = user_roles.target_id
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_id = sqlc.arg(team_id)::UUID AND users.service_account = false
ORDER BY users.name ASC;

-- name: GetTeamMetadata :many
SELECT * FROM team_metadata
WHERE team_id = $1;

-- name: SetTeamMetadata :exec
INSERT INTO team_metadata (team_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (team_id, key) DO
    UPDATE SET value = $3;

-- name: UpdateTeam :one
UPDATE teams
SET name = COALESCE(sqlc.narg(name), name), purpose = COALESCE(sqlc.arg(purpose), purpose)
WHERE id = sqlc.arg(id)
RETURNING *;
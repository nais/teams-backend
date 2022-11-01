-- name: CreateTeam :one
INSERT INTO teams (slug, purpose)
VALUES ($1, $2)
RETURNING *;

-- name: GetTeams :many
SELECT * FROM teams
ORDER BY slug ASC;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1;

-- name: GetTeamBySlug :one
SELECT * FROM teams
WHERE slug = $1;

-- name: GetTeamMembers :many
SELECT users.* FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1
ORDER BY users.name ASC;

-- name: GetTeamMetadata :many
SELECT * FROM team_metadata
WHERE team_slug = $1
ORDER BY key ASC;

-- name: SetTeamMetadata :exec
INSERT INTO team_metadata (team_slug, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (team_slug, key) DO
    UPDATE SET value = $3;

-- name: UpdateTeam :one
UPDATE teams
SET purpose = COALESCE(sqlc.narg(purpose), purpose)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DisableTeam :one
UPDATE teams
SET enabled = false
WHERE id = $1
RETURNING *;

-- name: EnableTeam :one
UPDATE teams
SET enabled = true
WHERE id = $1
RETURNING *;

-- name: RemoveUserFromTeam :exec
DELETE FROM user_roles
WHERE user_id = $1 AND target_team_slug = $2;
-- name: CreateTeam :one
INSERT INTO teams (id, name, slug, purpose) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1 LIMIT 1;

-- name: GetTeamBySlug :one
SELECT * FROM teams WHERE slug = $1 LIMIT 1;

-- name: GetTeams :many
SELECT * FROM teams ORDER BY name ASC;

-- name: GetTeamMembers :many
SELECT users.* FROM user_teams
JOIN users ON users.id = user_teams.user_id
WHERE user_teams.team_id = $1
ORDER BY users.name ASC;
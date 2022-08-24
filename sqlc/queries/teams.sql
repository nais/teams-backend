-- name: CreateTeam :one
INSERT INTO teams (id, name, slug, purpose) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTeam :one
UPDATE teams SET name = COALESCE(sqlc.narg('name'), name), purpose = COALESCE(sqlc.arg('purpose'), purpose) WHERE id = sqlc.arg('id') RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1 LIMIT 1;

-- name: GetTeamBySlug :one
SELECT * FROM teams WHERE slug = $1 LIMIT 1;

-- name: GetTeams :many
SELECT * FROM teams ORDER BY name ASC;

-- name: GetTeamMembers :many
SELECT users.* FROM user_roles
JOIN teams ON teams.id = user_roles.target_id
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_id = sqlc.arg(team_id)::UUID
ORDER BY users.name ASC;

-- name: AddTeamMember :exec
INSERT INTO user_roles (user_id, role_name, target_id) VALUES ($1, 'Team member', sqlc.arg(team_id)::UUID) ON CONFLICT DO NOTHING;

-- name: AddTeamOwner :exec
INSERT INTO user_roles (user_id, role_name, target_id) VALUES ($1, 'Team owner', sqlc.arg(team_id)::UUID) ON CONFLICT DO NOTHING;

-- name: RemoveUserFromTeam :exec
DELETE FROM user_roles WHERE user_id = $1 AND target_id = sqlc.arg(team_id)::UUID;
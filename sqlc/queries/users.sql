-- name: CreateUser :one
INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY name ASC;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUsersByEmail :many
SELECT * FROM users
WHERE email LIKE $1 LIMIT 1;

-- name: GetUserByApiKey :one
SELECT users.* FROM api_keys
JOIN users ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1 LIMIT 1;

-- name: CreateUserRole :exec
INSERT INTO user_roles (user_id, role_name, target_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: AddUserToTeam :exec
INSERT INTO user_teams (user_id, team_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: GetUserTeams :many
SELECT teams.* FROM user_teams JOIN teams ON teams.id = user_teams.team_id WHERE user_id = $1 ORDER BY teams.name ASC;

-- name: RemoveUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: SetUserName :one
UPDATE users SET name = $1 WHERE id = $2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
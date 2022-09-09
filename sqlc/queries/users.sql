-- name: CreateUser :one
INSERT INTO users (id, name, email, service_account) VALUES ($1, $2, sqlc.arg(email)::TEXT, false)
RETURNING *;

-- name: CreateServiceAccount :one
INSERT INTO users (id, name, service_account) VALUES (gen_random_uuid(), $1, true)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY name ASC;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetServiceAccount :one
SELECT * FROM users
WHERE name = $1 AND service_account = true LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = sqlc.arg(email)::TEXT LIMIT 1;

-- name: GetUsersByEmail :many
SELECT * FROM users
WHERE email LIKE sqlc.arg(email)::TEXT LIMIT 1;

-- name: GetUserByApiKey :one
SELECT users.* FROM api_keys
JOIN users ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1 LIMIT 1;

-- name: GetUserTeams :many
SELECT teams.* FROM user_roles JOIN teams ON teams.id = user_roles.target_id WHERE user_roles.user_id = $1 ORDER BY teams.name ASC;

-- name: GetUserRoles :many
SELECT * FROM user_roles
WHERE user_id = $1;

-- name: AssignGlobalRoleToUser :exec
INSERT INTO user_roles (user_id, role_name) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RevokeGlobalRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_name = $2;

-- name: AssignTargetedRoleToUser :exec
INSERT INTO user_roles (user_id, role_name, target_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: RevokeTargetedRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND target_id = $2 AND role_name = $3;

-- name: RemoveGlobalUserRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND target_id IS NULL AND role_name = $2;

-- name: RemoveAllUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: SetUserName :one
UPDATE users SET name = $1 WHERE id = $2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
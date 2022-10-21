-- name: CreateUser :one
INSERT INTO users (name, email, service_account, external_id)
VALUES ($1, LOWER(sqlc.arg(email)::TEXT), false, sqlc.arg(external_id)::TEXT)
RETURNING *;

-- name: CreateServiceAccount :one
INSERT INTO users (name, service_account)
VALUES ($1, true)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users
WHERE service_account = false
ORDER BY name ASC;

-- name: GetServiceAccounts :many
SELECT * FROM users
WHERE service_account = true
ORDER BY name ASC;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND service_account = false;

-- name: GetUserByExternalID :one
SELECT * FROM users
WHERE external_id = sqlc.arg(external_id)::TEXT AND service_account = false;

-- name: GetServiceAccountByName :one
SELECT * FROM users
WHERE name = $1 AND service_account = true;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = LOWER(sqlc.arg(email)::TEXT) AND service_account = false;

-- name: GetServiceAccountByApiKey :one
SELECT users.* FROM api_keys
JOIN users ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1 AND users.service_account = true;

-- name: GetUserTeams :many
SELECT teams.* FROM user_roles
JOIN teams ON teams.id = user_roles.target_id
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.user_id = $1 AND users.service_account = false
ORDER BY teams.slug ASC;

-- name: UpdateUser :one
UPDATE users
SET name = $1, email = LOWER(sqlc.arg(email)::TEXT), external_id = sqlc.arg(external_id)::TEXT
WHERE id = $2 AND service_account = false
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1 AND service_account = false;

-- name: DeleteServiceAccount :exec
DELETE FROM users
WHERE id = $1 AND service_account = true;
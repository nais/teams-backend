-- name: CreateUser :one
INSERT INTO users (name, email, external_id)
VALUES ($1, LOWER(sqlc.arg(email)), sqlc.arg(external_id))
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY name ASC;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByExternalID :one
SELECT * FROM users
WHERE external_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = LOWER(sqlc.arg(email));

-- name: GetUserTeams :many
SELECT teams.* FROM user_roles
JOIN teams ON teams.id = user_roles.target_id
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.user_id = $1
ORDER BY teams.slug ASC;

-- name: UpdateUser :one
UPDATE users
SET name = $1, email = LOWER(sqlc.arg(email)), external_id = $2
WHERE id = $3
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
-- name: CreateUser :one
INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: AddRoleToUser :exec
INSERT INTO user_roles (id, user_id, role_name, target_id) VALUES ($1, $2, $3, $4);

-- name: AddUserToTeam :exec
INSERT INTO user_teams (id, user_id, team_id) VALUES ($1, $2, $3);

-- name: GetUserTeams :many
SELECT * FROM user_teams WHERE user_id = $1;
-- name: AddRoleToUser :exec
INSERT INTO user_roles (id, user_id, role_id, created_by_id) VALUES ($1, $2, $3, $4);
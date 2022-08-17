-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1 LIMIT 1;

-- name: GetRoles :many
SELECT * FROM roles
ORDER BY name ASC;

-- name: GetUserRoles :many
SELECT * FROM user_roles
WHERE user_id = $1;

-- name: GetUserRole :one
SELECT * FROM user_roles
WHERE id = $1 LIMIT 1;
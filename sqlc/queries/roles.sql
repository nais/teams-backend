-- name: GetUserRoles :many
SELECT * FROM user_roles
WHERE user_id = $1;

-- name: GetUserRole :one
SELECT * FROM user_roles
WHERE id = $1 LIMIT 1;

-- name: GetRoleAuthorizations :many
SELECT authz_name
FROM role_authz
WHERE role_name = $1
ORDER BY authz_name ASC;

-- name: GetRoleNames :many
SELECT unnest(enum_range(NULL::role_name))::role_name;
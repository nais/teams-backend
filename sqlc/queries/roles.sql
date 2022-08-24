-- name: GetRoleAuthorizations :many
SELECT authz_name
FROM role_authz
WHERE role_name = $1
ORDER BY authz_name ASC;
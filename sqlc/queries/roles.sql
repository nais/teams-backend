-- name: GetRoleAuthorizations :many
SELECT authz_name FROM role_authz
WHERE role_name = $1
ORDER BY authz_name ASC;

-- name: GetUserRoles :many
SELECT * FROM user_roles
WHERE user_id = $1;

-- name: AssignGlobalRoleToUser :exec
INSERT INTO user_roles (user_id, role_name)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RevokeGlobalRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_name = $2;

-- name: AssignTargetedRoleToUser :exec
INSERT INTO user_roles (user_id, role_name, target_id)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: RevokeTargetedRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND target_id = $2 AND role_name = $3;

-- name: RemoveGlobalUserRole :exec
DELETE FROM user_roles
WHERE user_id = $1 AND target_id IS NULL AND role_name = $2;

-- name: RemoveAllUserRoles :exec
DELETE FROM user_roles
WHERE user_id = $1;
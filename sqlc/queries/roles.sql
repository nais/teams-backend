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

-- name: AssignGlobalRoleToServiceAccount :exec
INSERT INTO service_account_roles (service_account_id, role_name)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: AssignTeamRoleToUser :exec
INSERT INTO user_roles (user_id, role_name, target_team_slug)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: RemoveGlobalUserRole :exec
DELETE FROM user_roles
WHERE user_id = $1
AND target_team_slug IS NULL
AND target_service_account_id IS NULL
AND role_name = $2;

-- name: RemoveAllUserRoles :exec
DELETE FROM user_roles
WHERE user_id = $1;

-- name: RemoveAllServiceAccountRoles :exec
DELETE FROM service_account_roles
WHERE service_account_id = $1;
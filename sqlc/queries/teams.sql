-- name: CreateTeam :one
INSERT INTO teams (slug, purpose, slack_channel)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeams :many
SELECT teams.* FROM teams
ORDER BY teams.slug ASC;

-- name: GetTeamBySlug :one
SELECT teams.* FROM teams
WHERE teams.slug = $1;

-- name: GetTeamMembers :many
SELECT users.* FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1
ORDER BY users.name ASC;

-- name: GetTeamMember :one
SELECT users.* FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1 AND users.id = $2
ORDER BY users.name ASC;

-- name: GetTeamMembersForReconciler :many
SELECT users.* FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE
    user_roles.target_team_slug = $1
    AND user_roles.user_id NOT IN (
        SELECT user_id
        FROM reconciler_opt_outs
        WHERE team_slug = $1 AND reconciler_name = $2
    )
ORDER BY users.name ASC;

-- name: UpdateTeam :one
UPDATE teams
SET purpose = COALESCE(sqlc.narg(purpose), purpose),
    slack_channel = COALESCE(sqlc.narg(slack_channel), slack_channel)
WHERE slug = sqlc.arg(slug)
RETURNING *;

-- name: RemoveUserFromTeam :exec
DELETE FROM user_roles
WHERE user_id = $1 AND target_team_slug = $2;

-- name: SetLastSuccessfulSyncForTeam :exec
UPDATE teams SET last_successful_sync = NOW()
WHERE slug = $1;

-- name: GetSlackAlertsChannels :many
SELECT * FROM slack_alerts_channels
WHERE team_slug = $1
ORDER BY environment ASC;

-- name: SetSlackAlertsChannel :exec
INSERT INTO slack_alerts_channels (team_slug, environment, channel_name)
VALUES ($1, $2, $3)
ON CONFLICT (team_slug, environment) DO
    UPDATE SET channel_name = $3;

-- name: RemoveSlackAlertsChannel :exec
DELETE FROM slack_alerts_channels
WHERE team_slug = $1 AND environment = $2;

-- name: CreateTeamDeleteKey :one
INSERT INTO team_delete_keys (team_slug, created_by)
VALUES($1, $2)
RETURNING *;

-- name: GetTeamDeleteKey :one
SELECT * FROM team_delete_keys
WHERE key = $1;

-- name: ConfirmTeamDeleteKey :exec
UPDATE team_delete_keys
SET confirmed_at = NOW()
WHERE key = $1;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE slug = $1;

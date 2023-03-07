-- name: CreateTeam :one
INSERT INTO teams (slug, purpose, slack_channel)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeams :many
SELECT teams.* FROM teams
LEFT JOIN team_delete_keys ON team_delete_keys.team_slug = teams.slug
WHERE team_delete_keys.confirmed_at IS NULL
ORDER BY teams.slug ASC;

-- name: GetTeamBySlug :one
SELECT teams.* FROM teams
LEFT JOIN team_delete_keys ON team_delete_keys.team_slug = teams.slug
WHERE teams.slug = $1 AND team_delete_keys.confirmed_at IS NULL;

-- name: GetTeamMembers :many
SELECT users.* FROM user_roles
JOIN teams ON teams.slug = user_roles.target_team_slug
JOIN users ON users.id = user_roles.user_id
WHERE user_roles.target_team_slug = $1
ORDER BY users.name ASC;

-- name: GetTeamMetadata :many
SELECT * FROM team_metadata
WHERE team_slug = $1
ORDER BY key ASC;

-- name: SetTeamMetadata :exec
INSERT INTO team_metadata (team_slug, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (team_slug, key) DO
    UPDATE SET value = $3;

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
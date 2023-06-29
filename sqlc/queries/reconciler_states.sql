-- name: GetReconcilerStateForTeam :one
SELECT * FROM reconciler_states
WHERE reconciler = $1 AND team_slug = $2;

-- name: GetTeamsWithPermissionInRepo :many
select t.* from teams t
LEFT JOIN reconciler_states rs ON rs.team_slug = t.slug
    WHERE rs.reconciler = 'github:team'
    AND rs.state @> $1;

-- name: SetReconcilerStateForTeam :exec
INSERT INTO reconciler_states (reconciler, team_slug, state)
VALUES($1, $2, $3)
ON CONFLICT (reconciler, team_slug) DO
    UPDATE SET state = $3;

-- name: RemoveReconcilerStateForTeam :exec
DELETE FROM reconciler_states
WHERE reconciler = $1 AND team_slug = $2;
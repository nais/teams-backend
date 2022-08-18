-- name: AddUserToTeam :exec
INSERT INTO user_teams (id, user_id, team_id, created_by_id) VALUES ($1, $2, $3, $4);

-- name: GetUserTeams :many
SELECT * FROM user_teams WHERE user_id = $1;

-- name: GetTeamMembers :many
SELECT users.* FROM user_teams
JOIN users ON users.id = user_teams.user_id
WHERE user_teams.team_id = $1
ORDER BY users.name ASC;
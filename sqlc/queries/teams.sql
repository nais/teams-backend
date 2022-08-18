-- name: CreateTeam :one
INSERT INTO teams (id, name, slug, purpose, created_by_id) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTeam :one
SELECT * FROM teams WHERE id = $1 LIMIT 1;
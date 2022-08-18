-- name: GetSystem :one
SELECT * FROM systems
WHERE id = $1 LIMIT 1;

-- name: GetSystemByName :one
SELECT * FROM systems
WHERE name = $1 LIMIT 1;

-- name: GetSystems :many
SELECT * FROM systems
ORDER BY name ASC;

-- name: CreateSystem :one
INSERT INTO systems (id, name) VALUES ($1, $2)
RETURNING *;
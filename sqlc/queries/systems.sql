-- name: GetSystem :one
SELECT * FROM systems
WHERE id = $1 LIMIT 1;

-- name: GetSystems :many
SELECT * FROM systems
ORDER BY name ASC;
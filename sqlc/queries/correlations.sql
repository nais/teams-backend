-- name: CreateCorrelation :one
INSERT INTO correlations (id) VALUES ($1)
RETURNING *;
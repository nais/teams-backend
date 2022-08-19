-- name: CreateAPIKey :exec
INSERT INTO api_keys (api_key, user_id) VALUES ($1, $2);

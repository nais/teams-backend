-- name: CreateAPIKey :exec
INSERT INTO api_keys (api_key, user_id) VALUES ($1, $2);

-- name: RemoveApiKeysFromUser :exec
DELETE FROM api_keys WHERE user_id = $1;
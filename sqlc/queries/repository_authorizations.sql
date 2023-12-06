-- name: AddRepositoryAuthorization :one
INSERT INTO repository_authorizations (team_slug, github_repository, repository_authorization)
VALUES ($1, $2, $3)
RETURNING *;
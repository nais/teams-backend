-- name: CreateRepositoryAuthorization :exec
INSERT INTO repository_authorizations (team_slug, github_repository, repository_authorization)
VALUES ($1, $2, $3);

-- name: RemoveRepositoryAuthorization :exec
DELETE FROM repository_authorizations
WHERE
    team_slug = $1
    AND github_repository = $2
    AND repository_authorization = $3;
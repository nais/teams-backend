package db

import (
	"context"

	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
)

func (d *database) CreateRepositoryAuthorization(ctx context.Context, teamSlug slug.Slug, repoName string, authorization sqlc.RepositoryAuthorizationEnum) error {
	return d.querier.CreateRepositoryAuthorization(ctx, sqlc.CreateRepositoryAuthorizationParams{
		TeamSlug:                string(teamSlug),
		GithubRepository:        repoName,
		RepositoryAuthorization: authorization,
	})
}

func (d *database) RemoveRepositoryAuthorization(ctx context.Context, teamSlug slug.Slug, repoName string, authorization sqlc.RepositoryAuthorizationEnum) error {
	return d.querier.RemoveRepositoryAuthorization(ctx, sqlc.RemoveRepositoryAuthorizationParams{
		TeamSlug:                string(teamSlug),
		GithubRepository:        repoName,
		RepositoryAuthorization: authorization,
	})
}

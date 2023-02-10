package gar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	Name             = sqlc.ReconcilerNameGoogleGcpGar
	RepositoryFormat = "DOCKER"
)

type garReconciler struct {
	auditLogger             auditlogger.AuditLogger
	managementProjectID     string
	log                     logger.Logger
	artifactregistryService *artifactregistry.ProjectsLocationsRepositoriesService
}

func New(auditLogger auditlogger.AuditLogger, managementProjectID string, artifactregistryService *artifactregistry.ProjectsLocationsRepositoriesService, log logger.Logger) *garReconciler {
	return &garReconciler{
		auditLogger:             auditLogger,
		log:                     log,
		managementProjectID:     managementProjectID,
		artifactregistryService: artifactregistryService,
	}
}

func NewFromConfig(ctx context.Context, _ db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	builder, err := google_token_source.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	ts, err := builder.GCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("get delegated token source: %w", err)
	}

	baseService, err := artifactregistry.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	artifactregistryService := artifactregistry.NewProjectsLocationsRepositoriesService(baseService)

	return New(auditLogger, cfg.GoogleManagementProjectID, artifactregistryService, log), nil
}

func (r *garReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *garReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	slugStr := string(input.Team.Slug)
	repositoryName := fmt.Sprintf("projects/%s/locations/europe-north1/repositories/%s", r.managementProjectID, slugStr)

	existing, err := r.artifactregistryService.Get(repositoryName).Do()
	if googleError, ok := err.(*googleapi.Error); ok {
		if googleError.Code != http.StatusNotFound {
			return googleError
		}
	} else if err != nil {
		return err
	}

	template := &artifactregistry.Repository{
		Format:      RepositoryFormat,
		Name:        repositoryName,
		Description: fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", slugStr),
		Labels: map[string]string{
			"team": slugStr,
		},
	}

	if existing == nil {
		_, err := r.artifactregistryService.Create("projects/"+r.managementProjectID, template).Do()
		return err
	}

	if existing.Format != RepositoryFormat {
		return fmt.Errorf("existing repo has invalid format: %q %q", repositoryName, existing.Format)
	}

	hasChanges := r.ensureRepoConfig(existing, input.Team.Slug, template.Description)
	if !hasChanges {
		r.log.WithTeamSlug(slugStr).Debugf("repository exists and is correct")
		return nil
	}

	_, err = r.artifactregistryService.Patch(repositoryName, existing).Do()
	return err
}

func (r *garReconciler) ensureRepoConfig(repo *artifactregistry.Repository, slug slug.Slug, description string) (hasChanges bool) {
	if repo.Labels["team"] != string(slug) {
		repo.Labels["team"] = string(slug)
		hasChanges = true
	}

	if repo.Description != description {
		repo.Description = fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", slug)
		hasChanges = true
	}

	return
}

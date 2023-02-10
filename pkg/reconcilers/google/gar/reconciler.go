package gar

import (
	"context"
	"fmt"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	Name             = sqlc.ReconcilerNameGoogleGcpGar
	RepositoryFormat = "DOCKER"
)

type garReconciler struct {
	auditLogger         auditlogger.AuditLogger
	managementProjectID string
	log                 logger.Logger
	artifactRegistry    *artifactregistry.Client
}

func New(auditLogger auditlogger.AuditLogger, managementProjectID string, garClient *artifactregistry.Client, log logger.Logger) *garReconciler {
	return &garReconciler{
		auditLogger:         auditLogger,
		log:                 log,
		managementProjectID: managementProjectID,
		artifactRegistry:    garClient,
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

	garClient, err := artifactregistry.NewClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	return New(auditLogger, cfg.GoogleManagementProjectID, garClient, log), nil
}

func (r *garReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *garReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	slugStr := string(input.Team.Slug)
	parent := fmt.Sprintf("projects/%s/locations/europe-north1", r.managementProjectID)
	repositoryName := fmt.Sprintf("%s/repositories/%s", parent, slugStr)
	repositoryDescription := fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", slugStr)

	getRequest := &artifactregistrypb.GetRepositoryRequest{
		Name: repositoryName,
	}
	existing, err := r.artifactRegistry.GetRepository(ctx, getRequest)
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	if existing == nil {
		template := &artifactregistrypb.Repository{
			Format:      artifactregistrypb.Repository_DOCKER,
			Name:        repositoryName,
			Description: repositoryDescription,
			Labels: map[string]string{
				"team": slugStr,
			},
		}

		createRequest := &artifactregistrypb.CreateRepositoryRequest{
			Parent:     parent,
			Repository: template,
		}

		createResponse, err := r.artifactRegistry.CreateRepository(ctx, createRequest)
		if err != nil {
			return err
		}

		_, err = createResponse.Wait(ctx)
		return err
	}

	if existing.Format != artifactregistrypb.Repository_DOCKER {
		return fmt.Errorf("existing repo has invalid format: %q %q", repositoryName, existing.Format)
	}

	return r.updateRepository(ctx, existing, input.Team.Slug, repositoryDescription)
}

func (r *garReconciler) updateRepository(ctx context.Context, repo *artifactregistrypb.Repository, slug slug.Slug, description string) error {
	var changes []string
	if repo.Labels["team"] != string(slug) {
		repo.Labels["team"] = string(slug)
		changes = append(changes, "labels.team")
	}

	if repo.Description != description {
		repo.Description = fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", slug)
		changes = append(changes, "description")
	}

	if len(changes) > 0 {
		updateRequest := &artifactregistrypb.UpdateRepositoryRequest{
			Repository: repo,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: changes,
			},
		}

		_, err := r.artifactRegistry.UpdateRepository(ctx, updateRequest)
		return err
	}

	return nil
}

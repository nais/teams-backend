package google_gar

import (
	"context"
	"fmt"
	"net/http"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	Name = sqlc.ReconcilerNameGoogleGcpGar
)

type garReconciler struct {
	database                 db.Database
	auditLogger              auditlogger.AuditLogger
	managementProjectID      string
	workloadIdentityPoolName string
	log                      logger.Logger
	artifactRegistry         *artifactregistry.Client
	iamService               *iam.Service
}

func New(auditLogger auditlogger.AuditLogger, database db.Database, managementProjectID, workloadIdentityPoolName string, garClient *artifactregistry.Client, iamService *iam.Service, log logger.Logger) *garReconciler {
	return &garReconciler{
		database:                 database,
		auditLogger:              auditLogger,
		log:                      log,
		managementProjectID:      managementProjectID,
		workloadIdentityPoolName: workloadIdentityPoolName,
		artifactRegistry:         garClient,
		iamService:               iamService,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
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

	iamService, err := iam.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	return New(auditLogger, database, cfg.GoogleManagementProjectID, cfg.GCP.WorkloadIdentityPoolName, garClient, iamService, log), nil
}

func (r *garReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *garReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	log := r.log.WithTeamSlug(string(input.Team.Slug))
	serviceAccount, err := r.getOrCreateServiceAccount(ctx, input)
	if err != nil {
		return err
	}

	err = r.setServiceAccountPolicy(ctx, serviceAccount, input.Team.Slug)
	if err != nil {
		return err
	}

	garRepository, err := r.getOrCreateOrUpdateGarRepository(ctx, input, log)
	if err != nil {
		return err
	}

	err = r.setGarRepositoryPolicy(ctx, garRepository, serviceAccount)
	if err != nil {
		return err
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, reconcilers.GoogleGarState{
		RepositoryName: &garRepository.Name,
	})
	if err != nil {
		log.WithError(err).Error("persist reconciler state")
	}

	return nil
}

func (r *garReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	state := &reconcilers.GoogleGarState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	if state.RepositoryName == nil {
		return fmt.Errorf("missing repository name in reconciler state for team %q in reconciler %q", teamSlug, r.Name())
	}

	serviceAccountName, _ := serviceAccountNameAndAccountID(teamSlug, r.managementProjectID)
	_, err = r.iamService.Projects.ServiceAccounts.Delete(serviceAccountName).Context(ctx).Do()
	if err != nil {
		googleError, ok := err.(*googleapi.Error)
		if !ok || googleError.Code != http.StatusNotFound {
			return fmt.Errorf("delete service account %q: %w", serviceAccountName, err)
		}

		r.log.
			WithTeamSlug(string(teamSlug)).
			WithError(err).
			Infof("GAR service account %q does not exist, nothing to delete", serviceAccountName)
	}

	garRepositoryName := *state.RepositoryName

	req := &artifactregistrypb.DeleteRepositoryRequest{
		Name: garRepositoryName,
	}
	operation, err := r.artifactRegistry.DeleteRepository(ctx, req)
	if err != nil {
		return fmt.Errorf("delete GAR repository for team %q: %w", teamSlug, err)
	}

	err = operation.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for GAR repository deletion for team %q: %w", teamSlug, err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(teamSlug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleGarDelete,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Delete GAR repository %q", garRepositoryName)

	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func (r *garReconciler) getOrCreateServiceAccount(ctx context.Context, input reconcilers.Input) (*iam.ServiceAccount, error) {
	serviceAccountName, accountID := serviceAccountNameAndAccountID(input.Team.Slug, r.managementProjectID)

	existing, err := r.iamService.Projects.ServiceAccounts.Get(serviceAccountName).Context(ctx).Do()
	if err == nil {
		return existing, nil
	}

	return r.iamService.Projects.ServiceAccounts.Create("projects/"+r.managementProjectID, &iam.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iam.ServiceAccount{
			Description: "Service Account used to push images to Google Artifact Registry for " + string(input.Team.Slug),
			DisplayName: "Artifact Pusher for " + string(input.Team.Slug),
		},
	}).Context(ctx).Do()
}

func (r *garReconciler) setServiceAccountPolicy(ctx context.Context, serviceAccount *iam.ServiceAccount, teamSlug slug.Slug) error {
	members, err := r.getServiceAccountPolicyMembers(ctx, teamSlug)
	if err != nil {
		return err
	}

	req := iam.SetIamPolicyRequest{
		Policy: &iam.Policy{
			Bindings: []*iam.Binding{
				{
					Members: members,
					Role:    "roles/iam.workloadIdentityUser",
				},
			},
		},
	}

	_, err = r.iamService.Projects.ServiceAccounts.SetIamPolicy(serviceAccount.Name, &req).Context(ctx).Do()
	return err
}

func (r *garReconciler) getOrCreateOrUpdateGarRepository(ctx context.Context, input reconcilers.Input, log logger.Logger) (*artifactregistrypb.Repository, error) {
	parent := fmt.Sprintf("projects/%s/locations/europe-north1", r.managementProjectID)
	name := fmt.Sprintf("%s/repositories/%s", parent, input.Team.Slug)
	description := fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", input.Team.Slug)

	getRequest := &artifactregistrypb.GetRepositoryRequest{
		Name: name,
	}
	existing, err := r.artifactRegistry.GetRepository(ctx, getRequest)
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}

	if existing == nil {
		template := &artifactregistrypb.Repository{
			Format:      artifactregistrypb.Repository_DOCKER,
			Name:        name,
			Description: description,
			Labels: map[string]string{
				"team": string(input.Team.Slug),
			},
		}

		createRequest := &artifactregistrypb.CreateRepositoryRequest{
			Parent:       parent,
			Repository:   template,
			RepositoryId: string(input.Team.Slug),
		}

		createResponse, err := r.artifactRegistry.CreateRepository(ctx, createRequest)
		if err != nil {
			return nil, err
		}

		return createResponse.Wait(ctx)
	}

	if existing.Format != artifactregistrypb.Repository_DOCKER {
		return nil, fmt.Errorf("existing repo has invalid format: %q %q", name, existing.Format)
	}

	return r.updateGarRepository(ctx, existing, input.Team.Slug, description, log)
}

func (r *garReconciler) getServiceAccountPolicyMembers(ctx context.Context, teamSlug slug.Slug) ([]string, error) {
	state := reconcilers.GitHubState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGithubTeam, teamSlug, &state)
	if err != nil {
		return nil, err
	}

	members := make([]string, 0)
	for _, githubRepo := range state.Repositories {
		for _, perm := range githubRepo.Permissions {
			if perm.Name == "push" && perm.Granted {
				member := fmt.Sprintf("principalSet://iam.googleapis.com/%s/attribute.repository/%s", r.workloadIdentityPoolName, githubRepo.Name)
				members = append(members, member)
				break
			}
		}
	}

	return members, nil
}

func (r *garReconciler) updateGarRepository(ctx context.Context, repository *artifactregistrypb.Repository, slug slug.Slug, description string, log logger.Logger) (*artifactregistrypb.Repository, error) {
	var changes []string
	if repository.Labels["team"] != string(slug) {
		repository.Labels["team"] = string(slug)
		changes = append(changes, "labels.team")
	}

	if repository.Description != description {
		repository.Description = fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", slug)
		changes = append(changes, "description")
	}

	if len(changes) > 0 {
		updateRequest := &artifactregistrypb.UpdateRepositoryRequest{
			Repository: repository,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: changes,
			},
		}

		return r.artifactRegistry.UpdateRepository(ctx, updateRequest)
	}

	log.Debugf("existing repository is up to date")
	return repository, nil
}

func (r *garReconciler) setGarRepositoryPolicy(ctx context.Context, repository *artifactregistrypb.Repository, serviceAccount *iam.ServiceAccount) error {
	_, err := r.artifactRegistry.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: repository.Name,
		Policy: &iampb.Policy{
			Bindings: []*iampb.Binding{
				{
					Role:    "roles/artifactregistry.writer",
					Members: []string{"serviceAccount:" + serviceAccount.Email},
				},
			},
		},
	})
	return err
}

func serviceAccountNameAndAccountID(teamSlug slug.Slug, projectID string) (serviceAccountName, accountID string) {
	accountId := console.SlugHashPrefixTruncate(teamSlug, "gar", gcp.GoogleServiceAccountMaxLength)
	emailAddress := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountId, projectID)
	serviceAccountName = fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, emailAddress)
	return
}

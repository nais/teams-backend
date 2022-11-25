package nais_namespace_reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
)

const (
	NaisdCreateNamespace = "create-namespace"
)

type naisdData struct {
	Name         string `json:"name"`
	GcpProject   string `json:"gcpProject"` // the user specified "project id"; not the "projects/ID" format
	GroupEmail   string `json:"groupEmail"`
	AzureGroupID string `json:"azureGroupID"`
}

type naisdRequest struct {
	Type string    `json:"type"`
	Data naisdData `json:"data"`
}

type naisNamespaceReconciler struct {
	database     db.Database
	domain       string
	auditLogger  auditlogger.AuditLogger
	projectID    string
	azureEnabled bool
	tokenSource  oauth2.TokenSource
	log          logger.Logger
}

const Name = sqlc.ReconcilerNameNaisNamespace

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain, projectID string, azureEnabled bool, tokenSource oauth2.TokenSource, log logger.Logger) *naisNamespaceReconciler {
	return &naisNamespaceReconciler{
		database:     database,
		auditLogger:  auditLogger,
		domain:       domain,
		projectID:    projectID,
		azureEnabled: azureEnabled,
		tokenSource:  tokenSource,
		log:          log,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	ts, err := google_token_source.NewFromConfig(cfg).GCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("create token source: %w", err)
	}
	return New(database, auditLogger, cfg.TenantDomain, cfg.GoogleManagementProjectID, cfg.NaisNamespace.AzureEnabled, ts, log), nil
}

func (r *naisNamespaceReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *naisNamespaceReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	svc, err := pubsub.NewClient(ctx, r.projectID, option.WithTokenSource(r.tokenSource))
	if err != nil {
		return fmt.Errorf("retrieve pubsub client: %w", err)
	}

	namespaceState := &reconcilers.GoogleGcpNaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err = r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, namespaceState)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	gcpProjectState := &reconcilers.GoogleGcpProjectState{}
	err = r.database.LoadReconcilerStateForTeam(ctx, google_gcp_reconciler.Name, input.Team.Slug, gcpProjectState)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, google_gcp_reconciler.Name, err)
	}

	if len(gcpProjectState.Projects) == 0 {
		return fmt.Errorf("no GCP project state exists for team %q yet", input.Team.Slug)
	}

	googleGroupEmail, err := r.getGoogleGroupEmail(ctx, input.Team.Slug)
	if err != nil {
		return err
	}

	azureGroupID, err := r.getAzureGroupID(ctx, input.Team.Slug)
	if err != nil {
		return err
	}

	for environment, project := range gcpProjectState.Projects {
		err = r.createNamespace(ctx, svc, input.Team, environment, project.ProjectID, googleGroupEmail, azureGroupID)
		if err != nil {
			return fmt.Errorf("unable to create namespace for project %q in environment %q: %w", project.ProjectID, environment, err)
		}

		if _, requested := namespaceState.Namespaces[environment]; !requested {
			targets := []auditlogger.Target{
				auditlogger.TeamTarget(input.Team.Slug),
			}
			fields := auditlogger.Fields{
				Action:        sqlc.AuditActionNaisNamespaceCreateNamespace,
				CorrelationID: input.CorrelationID,
			}
			r.auditLogger.Logf(ctx, targets, fields, "request namespace creation for team %q in environment %q", input.Team.Slug, environment)
			namespaceState.Namespaces[environment] = input.Team.Slug
		}
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, namespaceState)
	if err != nil {
		r.log.WithError(err).Error("persisted system state")
	}

	return nil
}

func (r *naisNamespaceReconciler) createNamespace(ctx context.Context, pubsubService *pubsub.Client, team db.Team, environment, gcpProjectID string, groupEmail string, azureGroupID string) error {
	const topicPrefix = "naisd-console-"
	req := &naisdRequest{
		Type: NaisdCreateNamespace,
		Data: naisdData{
			Name:         string(team.Slug),
			GcpProject:   gcpProjectID,
			GroupEmail:   groupEmail,
			AzureGroupID: azureGroupID,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	topic := topicPrefix + environment
	msg := &pubsub.Message{Data: payload}
	future := pubsubService.Topic(topic).Publish(ctx, msg)
	<-future.Ready()
	_, err = future.Get(ctx)
	return err
}

func (r *naisNamespaceReconciler) getGoogleGroupEmail(ctx context.Context, teamSlug slug.Slug) (string, error) {
	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, google_workspace_admin_reconciler.Name, teamSlug, googleWorkspaceState)
	if err != nil {
		return "", fmt.Errorf("no workspace admin state exists for team %q", teamSlug)
	}

	if googleWorkspaceState.GroupEmail == nil {
		return "", fmt.Errorf("no group email set for team %q", teamSlug)
	}

	return *googleWorkspaceState.GroupEmail, nil
}

func (r *naisNamespaceReconciler) getAzureGroupID(ctx context.Context, teamSlug slug.Slug) (string, error) {
	if !r.azureEnabled {
		return "", nil
	}

	azureState := &reconcilers.AzureState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, azure_group_reconciler.Name, teamSlug, azureState)
	if err != nil {
		return "", fmt.Errorf("no Azure state exists for team %q", teamSlug)
	}

	if azureState.GroupID == nil {
		return "", fmt.Errorf("no Azure group ID set for team %q", teamSlug)
	}

	return azureState.GroupID.String(), nil
}

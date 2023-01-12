package nais_namespace_reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/legacy/envmap"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"google.golang.org/api/option"
)

const (
	NaisdCreateNamespace = "create-namespace"
)

const metricsSystemName = "naisd"

type naisdData struct {
	Name               string `json:"name"`
	GcpProject         string `json:"gcpProject"` // the user specified "project id"; not the "projects/ID" format
	GroupEmail         string `json:"groupEmail"`
	AzureGroupID       string `json:"azureGroupID"`
	CNRMEmail          string `json:"cnrmEmail"`
	SlackAlertsChannel string `json:"slackAlertsChannel"`
}

type NaisdRequest struct {
	Type string    `json:"type"`
	Data naisdData `json:"data"`
}

type naisNamespaceReconciler struct {
	database       db.Database
	domain         string
	auditLogger    auditlogger.AuditLogger
	clusters       gcp.Clusters
	projectID      string
	azureEnabled   bool
	pubsubClient   *pubsub.Client
	log            logger.Logger
	legacyMapping  []envmap.EnvironmentMapping
	legacyClusters map[string]string
}

const Name = sqlc.ReconcilerNameNaisNamespace

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters gcp.Clusters, domain, projectID string, azureEnabled bool, pubsubClient *pubsub.Client, legacyMapping []envmap.EnvironmentMapping, legacyClusters map[string]string, log logger.Logger) *naisNamespaceReconciler {
	return &naisNamespaceReconciler{
		database:       database,
		auditLogger:    auditLogger,
		clusters:       clusters,
		domain:         domain,
		projectID:      projectID,
		azureEnabled:   azureEnabled,
		pubsubClient:   pubsubClient,
		legacyMapping:  legacyMapping,
		legacyClusters: legacyClusters,
		log:            log,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	builder, err := google_token_source.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	tokenSource, err := builder.GCP(ctx)
	if err != nil {
		return nil, fmt.Errorf("create token source: %w", err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, cfg.GoogleManagementProjectID, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("retrieve pubsub client: %w", err)
	}

	return New(database, auditLogger, cfg.GCP.Clusters, cfg.TenantDomain, cfg.GoogleManagementProjectID, cfg.NaisNamespace.AzureEnabled, pubsubClient, cfg.LegacyNaisNamespaces, cfg.LegacyClusters, log), nil
}

func (r *naisNamespaceReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func GCPProjectsWithLegacyEnvironments(projects map[string]reconcilers.GoogleGcpEnvironmentProject, mappings []envmap.EnvironmentMapping) map[string]reconcilers.GoogleGcpEnvironmentProject {
	output := make(map[string]reconcilers.GoogleGcpEnvironmentProject)
	for _, mapping := range mappings {
		output[mapping.Virtual] = projects[mapping.Real]
	}
	for k, v := range projects {
		output[k] = v
	}
	return output
}

func (r *naisNamespaceReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	namespaceState := &reconcilers.GoogleGcpNaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, namespaceState)
	if err != nil {
		return fmt.Errorf("unable to load NAIS namespace state for team %q: %w", input.Team.Slug, err)
	}

	gcpProjectState := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err = r.database.LoadReconcilerStateForTeam(ctx, google_gcp_reconciler.Name, input.Team.Slug, gcpProjectState)
	if err != nil {
		return fmt.Errorf("unable to load GCP project state for team %q: %w", input.Team.Slug, err)
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

	log := r.log.WithTeamSlug(string(input.Team.Slug))
	updateGcpProjectState := false

	// lag et merged array med environments

	projects := GCPProjectsWithLegacyEnvironments(gcpProjectState.Projects, r.legacyMapping)

	for environment, project := range projects {
		if !r.activeEnvironment(environment) {
			updateGcpProjectState = true
			log.Infof("environment %q from GCP project state is no longer active, will update state for the team", environment)
			delete(gcpProjectState.Projects, environment)
			continue
		}
		err = r.createNamespace(ctx, input.Team, environment, project.ProjectID, googleGroupEmail, azureGroupID)
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
			r.auditLogger.Logf(ctx, r.database, targets, fields, "Request namespace creation for team %q in environment %q", input.Team.Slug, environment)
			namespaceState.Namespaces[environment] = input.Team.Slug
		}
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, namespaceState)
	if err != nil {
		log.WithError(err).Error("persisted NAIS namespace state")
	}

	if updateGcpProjectState {
		err = r.database.SetReconcilerStateForTeam(ctx, google_gcp_reconciler.Name, input.Team.Slug, gcpProjectState)
		if err != nil {
			log.WithError(err).Error("persisted GCP project state")
		}
	}

	return nil
}

func (r *naisNamespaceReconciler) getClusterProjectForEnv(environment string) string {
	if cluster, ok := r.clusters[environment]; ok {
		return cluster.ProjectID
	}

	// fallback to legacy clusters
	return r.legacyClusters[environment]
}

func (r *naisNamespaceReconciler) createNamespace(ctx context.Context, team db.Team, environment, gcpProjectID string, groupEmail string, azureGroupID string) error {
	const topicPrefix = "naisd-console-"

	clusterProjectID := r.getClusterProjectForEnv(environment)
	cnrmAccountName, _ := google_gcp_reconciler.CnrmServiceAccountNameAndAccountID(team.Slug, clusterProjectID)
	parts := strings.Split(cnrmAccountName, "/")
	cnrmEmail := parts[len(parts)-1]

	req := &NaisdRequest{
		Type: NaisdCreateNamespace,
		Data: naisdData{
			Name:               string(team.Slug),
			GcpProject:         gcpProjectID,
			GroupEmail:         groupEmail,
			AzureGroupID:       azureGroupID,
			CNRMEmail:          cnrmEmail,
			SlackAlertsChannel: team.SlackChannel,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	topicName := topicPrefix + environment
	msg := &pubsub.Message{Data: payload}
	topic := r.pubsubClient.Topic(topicName)
	future := topic.Publish(ctx, msg)
	<-future.Ready()
	_, err = future.Get(ctx)
	topic.Stop()

	metrics.IncExternalCallsByError(metricsSystemName, err)

	return err
}

func (r *naisNamespaceReconciler) getGoogleGroupEmail(ctx context.Context, teamSlug slug.Slug) (string, error) {
	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, google_workspace_admin_reconciler.Name, teamSlug, googleWorkspaceState)
	if err != nil {
		return "", fmt.Errorf("no workspace admin state exists for team %q: %w", teamSlug, err)
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
		return "", fmt.Errorf("no Azure state exists for team %q: %w", teamSlug, err)
	}

	if azureState.GroupID == nil {
		return "", fmt.Errorf("no Azure group ID set for team %q", teamSlug)
	}

	return azureState.GroupID.String(), nil
}

func (r *naisNamespaceReconciler) activeEnvironment(environment string) bool {
	_, exists := r.clusters[environment]
	if exists {
		return true
	}
	for _, mapping := range r.legacyMapping {
		if mapping.Virtual == environment {
			_, exists = r.clusters[mapping.Real]
			return exists
		}
	}
	return false
}

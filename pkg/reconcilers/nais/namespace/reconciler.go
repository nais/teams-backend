package nais_namespace_reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nais/teams-backend/pkg/types"

	"cloud.google.com/go/pubsub"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/google_token_source"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/teams-backend/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"google.golang.org/api/option"
)

const (
	metricsSystemName        = "naisd"
	NaisdTypeCreateNamespace = "create-namespace"
	NaisdTypeDeleteNamespace = "delete-namespace"
	Name                     = sqlc.ReconcilerNameNaisNamespace
)

type NaisdCreateNamespace struct {
	Name               string `json:"name"`
	GcpProject         string `json:"gcpProject"` // the user specified "project id"; not the "projects/ID" format
	GroupEmail         string `json:"groupEmail"`
	AzureGroupID       string `json:"azureGroupID"`
	CNRMEmail          string `json:"cnrmEmail"`
	SlackAlertsChannel string `json:"slackAlertsChannel"`
}

type NaisdDeleteNamespace struct {
	Name string `json:"name"`
}

type NaisdRequest struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
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
	onpremClusters []string
}

func New(database db.Database, auditLogger auditlogger.AuditLogger, clusters gcp.Clusters, domain, projectID string, azureEnabled bool, pubsubClient *pubsub.Client, log logger.Logger, onpremClusters []string) *naisNamespaceReconciler {
	return &naisNamespaceReconciler{
		database:       database,
		auditLogger:    auditLogger,
		clusters:       clusters,
		domain:         domain,
		projectID:      projectID,
		azureEnabled:   azureEnabled,
		pubsubClient:   pubsubClient,
		log:            log.WithComponent(types.ComponentNameNaisNamespace),
		onpremClusters: onpremClusters,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
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

	return New(database, auditlogger.New(database, types.ComponentNameNaisNamespace, log), cfg.GCP.Clusters, cfg.TenantDomain, cfg.GoogleManagementProjectID, cfg.NaisNamespace.AzureEnabled, pubsubClient, log, cfg.OnpremClusters), nil
}

func (r *naisNamespaceReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *naisNamespaceReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	namespaceState := &reconcilers.NaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, namespaceState)
	if err != nil {
		return fmt.Errorf("unable to load NAIS namespace state for team %q: %w", input.Team.Slug, err)
	}
	if namespaceState.Namespaces == nil {
		namespaceState.Namespaces = make(map[string]slug.Slug)
	}

	gcpProjectState := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	err = r.database.LoadReconcilerStateForTeam(ctx, google_gcp_reconciler.Name, input.Team.Slug, gcpProjectState)
	if err != nil {
		return fmt.Errorf("unable to load GCP project state for team %q: %w", input.Team.Slug, err)
	}
	if gcpProjectState.Projects == nil {
		gcpProjectState.Projects = make(map[string]reconcilers.GoogleGcpEnvironmentProject)
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

	slackAlertsChannels, err := r.database.GetSlackAlertsChannels(ctx, input.Team.Slug)
	if err != nil {
		return err
	}

	projects := gcpProjectState.Projects
	for _, cluster := range r.onpremClusters {
		projects[cluster] = reconcilers.GoogleGcpEnvironmentProject{ProjectID: ""}
	}

	for environment, project := range projects {
		if !r.activeEnvironment(environment) {
			updateGcpProjectState = true
			log.Infof("environment %q from GCP project state is no longer active, will update state for the team", environment)
			delete(gcpProjectState.Projects, environment)
			continue
		}

		slackAlertsChannel := input.Team.SlackChannel
		if channel, exists := slackAlertsChannels[environment]; exists {
			slackAlertsChannel = channel
		}

		err = r.createNamespace(ctx, input.Team, environment, slackAlertsChannel, project.ProjectID, googleGroupEmail, azureGroupID)
		if err != nil {
			return fmt.Errorf("unable to create namespace for project %q in environment %q: %w", project.ProjectID, environment, err)
		}

		if _, requested := namespaceState.Namespaces[environment]; !requested {
			targets := []auditlogger.Target{
				auditlogger.TeamTarget(input.Team.Slug),
			}
			fields := auditlogger.Fields{
				Action:        types.AuditActionNaisNamespaceCreateNamespace,
				CorrelationID: input.CorrelationID,
			}
			r.auditLogger.Logf(ctx, targets, fields, "Request namespace creation for team %q in environment %q", input.Team.Slug, environment)
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

func (r *naisNamespaceReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	log := r.log.WithTeamSlug(teamSlug.String())
	namespaceState := &reconcilers.NaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, namespaceState)
	if err != nil {
		return fmt.Errorf("unable to load NAIS namespace state for team %q: %w", teamSlug, err)
	}

	if len(namespaceState.Namespaces) == 0 {
		r.log.Warnf("no namespaces for team %q in reconciler %q, assume already deleted", teamSlug, r.Name())
		return nil
	}

	var errors []error
	for environment := range namespaceState.Namespaces {
		if !r.activeEnvironment(environment) {
			log.Infof("environment %q from namespace state is no longer active, will update state for the team", environment)
			delete(namespaceState.Namespaces, environment)
			continue
		}

		if err := r.deleteNamespace(ctx, teamSlug, environment, correlationID); err != nil {
			log.WithError(err).Error("delete namespace")
			errors = append(errors, err)
		} else {
			targets := []auditlogger.Target{auditlogger.TeamTarget(teamSlug)}
			fields := auditlogger.Fields{
				Action:        types.AuditActionNaisNamespaceDeleteNamespace,
				CorrelationID: correlationID,
			}

			r.auditLogger.Logf(ctx, targets, fields, "Request namespace deletion for team %q in environment %q", teamSlug, environment)
			delete(namespaceState.Namespaces, environment)
		}
	}

	if len(errors) == 0 {
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), teamSlug, namespaceState)
	if err != nil {
		log.WithError(err).Error("set reconciler state")
	}

	return fmt.Errorf("%d errors occured during namespace deletion", len(errors))
}

func (r *naisNamespaceReconciler) deleteNamespace(ctx context.Context, teamSlug slug.Slug, environment string, correlationID uuid.UUID) error {
	const topicPrefix = "naisd-console-"

	deleteReq, err := json.Marshal(NaisdDeleteNamespace{
		Name: string(teamSlug),
	})
	if err != nil {
		return fmt.Errorf("marshal delete namespace request: %w", err)
	}

	payload, err := json.Marshal(NaisdRequest{
		Type: NaisdTypeDeleteNamespace,
		Data: deleteReq,
	})
	if err != nil {
		return fmt.Errorf("marshal naisd request envelope: %w", err)
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

func (r *naisNamespaceReconciler) createNamespace(ctx context.Context, team db.Team, environment, slackAlertsChannel, gcpProjectID, groupEmail, azureGroupID string) error {
	const topicPrefix = "naisd-console-"

	CNRMEmail := ""
	if gcpProjectID != "" {
		CNRMEmail = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", reconcilers.CnrmServiceAccountAccountID, gcpProjectID)
	}

	createReq, err := json.Marshal(
		NaisdCreateNamespace{
			Name:               string(team.Slug),
			GcpProject:         gcpProjectID,
			GroupEmail:         groupEmail,
			AzureGroupID:       azureGroupID,
			CNRMEmail:          CNRMEmail,
			SlackAlertsChannel: slackAlertsChannel,
		},
	)
	if err != nil {
		return fmt.Errorf("marshal create namespace request: %w", err)
	}

	payload, err := json.Marshal(NaisdRequest{
		Type: NaisdTypeCreateNamespace,
		Data: createReq,
	})
	if err != nil {
		return fmt.Errorf("marshal naisd request envelope: %w", err)
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

	if azureState.GroupID == uuid.Nil {
		return "", fmt.Errorf("no Azure group ID set for team %q", teamSlug)
	}

	return azureState.GroupID.String(), nil
}

func (r *naisNamespaceReconciler) activeEnvironment(environment string) bool {
	_, exists := r.clusters[environment]
	if exists {
		return true
	}
	for _, cluster := range r.onpremClusters {
		if cluster == environment {
			return true
		}
	}
	return false
}

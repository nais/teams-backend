package nais_namespace_reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"

	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"

	"cloud.google.com/go/pubsub"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"google.golang.org/api/option"
)

const (
	NaisdCreateNamespace = "create-namespace"
)

type naisdData struct {
	Name       string `json:"name"`
	GcpProject string `json:"gcpProject"` // the user specified "project id"; not the "projects/ID" format
}

type naisdRequest struct {
	Type string    `json:"type"`
	Data naisdData `json:"data"`
}

type naisNamespaceReconciler struct {
	database        db.Database
	domain          string
	auditLogger     auditlogger.AuditLogger
	credentialsFile string
	projectID       string
}

const Name = sqlc.SystemNameNaisNamespace

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain, credentialsFile, projectID string) *naisNamespaceReconciler {
	return &naisNamespaceReconciler{
		database:        database,
		auditLogger:     auditLogger,
		domain:          domain,
		credentialsFile: credentialsFile,
		projectID:       projectID,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.NaisNamespace.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	return New(database, auditLogger, cfg.TenantDomain, cfg.Google.CredentialsFile, cfg.NaisNamespace.ProjectID), nil
}

func (r *naisNamespaceReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *naisNamespaceReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	svc, err := pubsub.NewClient(ctx, r.projectID, option.WithCredentialsFile(r.credentialsFile))
	if err != nil {
		return fmt.Errorf("retrieve pubsub client: %w", err)
	}

	namespaceState := &reconcilers.GoogleGcpNaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	err = r.database.LoadSystemState(ctx, r.Name(), input.Team.ID, namespaceState)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	gcpProjectState := &reconcilers.GoogleGcpProjectState{}
	err = r.database.LoadSystemState(ctx, google_gcp_reconciler.Name, input.Team.ID, gcpProjectState)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, google_gcp_reconciler.Name, err)
	}

	if len(gcpProjectState.Projects) == 0 {
		return fmt.Errorf("no GCP project state exists for team %q yet", input.Team.Slug)
	}

	for environment, project := range gcpProjectState.Projects {
		err = r.createNamespace(ctx, svc, input.Team, environment, project.ProjectID)
		if err != nil {
			return fmt.Errorf("unable to create namespace for project %q in environment %q: %w", project.ProjectID, environment, err)
		}

		if _, requested := namespaceState.Namespaces[environment]; !requested {
			fields := auditlogger.Fields{
				Action:         sqlc.AuditActionNaisNamespaceCreateNamespace,
				CorrelationID:  input.CorrelationID,
				TargetTeamSlug: &input.Team.Slug,
			}
			r.auditLogger.Logf(ctx, fields, "request namespace creation for team %q in environment %q", input.Team.Slug, environment)
			namespaceState.Namespaces[environment] = input.Team.Slug
		}
	}

	err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, namespaceState)
	if err != nil {
		log.Errorf("system state not persisted: %s", err)
	}

	return nil
}

func (r *naisNamespaceReconciler) createNamespace(ctx context.Context, pubsubService *pubsub.Client, team db.Team, environment, gcpProjectID string) error {
	const topicPrefix = "naisd-console-"
	req := &naisdRequest{
		Type: NaisdCreateNamespace,
		Data: naisdData{
			Name:       string(team.Slug),
			GcpProject: gcpProjectID,
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

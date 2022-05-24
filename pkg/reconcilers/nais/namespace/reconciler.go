package google_gcp_reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
	"gorm.io/gorm"
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

type namespaceReconciler struct {
	db               *gorm.DB
	config           *jwt.Config
	domain           string
	logger           auditlogger.Logger
	projectParentIDs map[string]string
	topicPrefix      string
}

const (
	Name              = "nais:namespace"
	OpCreateNamespace = "nais:namespace:create-namespace"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(db *gorm.DB, logger auditlogger.Logger, domain, topicPrefix string, config *jwt.Config, projectParentIDs map[string]string) *namespaceReconciler {
	return &namespaceReconciler{
		db:               db,
		logger:           logger,
		domain:           domain,
		topicPrefix:      topicPrefix,
		config:           config,
		projectParentIDs: projectParentIDs,
	}
}

func NewFromConfig(db *gorm.DB, cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	if !cfg.NaisNamespace.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	b, err := ioutil.ReadFile(cfg.NaisNamespace.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	cf, err := google.JWTConfigFromJSON(
		b,
		cloudresourcemanager.CloudPlatformScope,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	return New(db, logger, cfg.NaisNamespace.Domain, cfg.NaisNamespace.TopicPrefix, cf, cfg.GCP.ProjectParentIDs), nil
}

func (s *namespaceReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	client := s.config.Client(ctx)

	svc, err := pubsub.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve pubsub client: %s", err)
	}

	for environment := range s.projectParentIDs {
		err = s.createNamespace(svc, in.Team, environment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *namespaceReconciler) createNamespace(pubsubService *pubsub.Service, team *dbmodels.Team, environment string) error {
	meta := &dbmodels.TeamMetadata{}
	key := dbmodels.TeamMetaGoogleProjectID + ":" + environment
	tx := s.db.First(meta, "key = ? AND team_id = ?", key, team.ID)
	if tx.Error != nil {
		return tx.Error
	}

	if meta.Value == nil {
		return fmt.Errorf("no GCP project created for team '%s' and environment '%s'", team.Slug, environment)
	}

	req := &naisdRequest{
		Type: NaisdCreateNamespace,
		Data: naisdData{
			Name:       team.Slug.String(),
			GcpProject: *meta.Value,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	publishRequest := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{
			{
				Data: string(payload),
			},
		},
	}

	topic := s.topicPrefix + environment
	_, err = pubsubService.Projects.Topics.Publish(topic, publishRequest).Do()

	return err
}

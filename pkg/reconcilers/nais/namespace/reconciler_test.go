package nais_namespace_reconciler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestReconcile(t *testing.T) {
	const (
		domain              = "example.com"
		managementProjectID = "some-project-123"
		teamProjectID       = "some-project-id"
		teamSlug            = "slug"
		environment         = "dev"
	)

	ctx := context.Background()
	team := db.Team{Team: &sqlc.Team{Slug: teamSlug}}
	input := reconcilers.Input{
		CorrelationID: uuid.New(),
		Team:          team,
	}

	googleWorkspaceEmail := "group-email@example.com"
	azureEnabled := true
	azureGroupID := uuid.New()
	clusters := gcp.Clusters{
		environment: gcp.Cluster{
			TeamFolderID: 123,
			ProjectID:    "env-dev-123",
		},
	}

	t.Run("unable to load namespace state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameNaisNamespace, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load NAIS namespace state for team "slug"`)
	})

	t.Run("unable to load GCP project state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load GCP project state for team "slug"`)
	})

	t.Run("no GCP projects in state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no GCP project state exists for team "slug"`)
	})

	t.Run("unable to get google group email", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: "some-project-id",
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no workspace admin state exists for team "slug"`)
	})

	t.Run("unable to get azure group id", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: "some-project-id",
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &googleWorkspaceEmail
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no Azure state exists for team "slug"`)
	})

	t.Run("create namespaces", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-dev")
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == sqlc.AuditActionNaisNamespaceCreateNamespace
			}), mock.Anything, team.Slug, environment).
			Return(nil).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: teamProjectID,
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &googleWorkspaceEmail
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.AzureState)
				state.GroupID = &azureGroupID
			}).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.GoogleGcpNaisNamespaceState) bool {
				return state.Namespaces[environment] == team.Slug
			})).
			Return(nil).
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log)
		assert.NoError(t, r.Reconcile(ctx, input))

		msgs := srv.Messages()
		assert.Len(t, msgs, 1)

		publishRequest := &nais_namespace_reconciler.NaisdRequest{}
		json.Unmarshal(msgs[0].Data, publishRequest)

		assert.Equal(t, teamSlug, publishRequest.Data.Name)
		assert.Equal(t, teamProjectID, publishRequest.Data.GcpProject)
		assert.Equal(t, googleWorkspaceEmail, publishRequest.Data.GroupEmail)
		assert.Equal(t, azureGroupID.String(), publishRequest.Data.AzureGroupID)
	})

	t.Run("environment in state no longer active", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-dev")
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
			Once()
		log.
			On("Infof", mock.MatchedBy(func(msg string) bool {
				return strings.Contains(msg, "from GCP project state is no longer active")
			}), environment).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects[environment] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: teamProjectID,
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &googleWorkspaceEmail
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.AzureState)
				state.GroupID = &azureGroupID
			}).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.GoogleGcpNaisNamespaceState) bool {
				return len(state.Namespaces) == 0
			})).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.GoogleGcpProjectState) bool {
				return len(state.Projects) == 0
			})).
			Return(nil).
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, gcp.Clusters{}, domain, managementProjectID, azureEnabled, pubsubClient, log)
		assert.NoError(t, r.Reconcile(ctx, input))

		msgs := srv.Messages()
		assert.Len(t, msgs, 0)
	})
}

func getPubsubServerAndClient(ctx context.Context, projectID string, topics ...string) (*pstest.Server, *pubsub.Client, func()) {
	srv := pstest.NewServer()
	client, _ := pubsub.NewClient(
		ctx,
		projectID,
		option.WithEndpoint(srv.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)

	for _, topic := range topics {
		client.CreateTopic(ctx, topic)
	}

	return srv, client, func() {
		srv.Close()
		client.Close()
	}
}

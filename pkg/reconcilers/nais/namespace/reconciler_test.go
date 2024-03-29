package nais_namespace_reconciler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/teams-backend/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/workspace_admin"
	nais_namespace_reconciler "github.com/nais/teams-backend/pkg/reconcilers/nais/namespace"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestReconcile(t *testing.T) {
	const (
		domain              = "example.com"
		managementProjectID = "management-project-123"
		teamProjectID       = "team-project-123"
		teamSlug            = "slug"
		environment         = "dev"
		clusterProjectID    = "cluster-dev-123"
		cnrmEmail           = "nais-sa-cnrm@team-project-123.iam.gserviceaccount.com"
		slackChannel        = "#team-channel"
	)

	ctx := context.Background()
	team := db.Team{Team: &sqlc.Team{Slug: teamSlug, SlackChannel: slackChannel}}
	input := reconcilers.Input{
		CorrelationID: uuid.New(),
		Team:          team,
	}

	googleWorkspaceEmail := "group-email@example.com"
	azureEnabled := true
	azureGroupID := uuid.New()
	clusters := gcp.Clusters{
		environment: gcp.Cluster{
			TeamsFolderID: 123,
			ProjectID:     clusterProjectID,
		},
	}
	noOnpremClusters := make([]string, 0)

	t.Run("unable to load namespace state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameNaisNamespace, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load NAIS namespace state for team "slug"`)
	})

	t.Run("unable to load GCP project state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
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
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load GCP project state for team "slug"`)
	})

	t.Run("no GCP projects in state", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
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
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no GCP project state exists for team "slug"`)
	})

	t.Run("unable to get google group email", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
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
					ProjectID: "some-project-id",
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no workspace admin state exists for team "slug"`)
	})

	t.Run("unable to get azure group id", func(t *testing.T) {
		_, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
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
		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no Azure state exists for team "slug"`)
	})

	t.Run("create namespaces", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-dev")
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
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
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.NaisNamespaceState) bool {
				return state.Namespaces[environment] == team.Slug
			})).
			Return(nil).
			Once()
		database.
			On("GetSlackAlertsChannels", ctx, team.Slug).
			Return(map[string]string{
				environment: "#env-channel",
			}, nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.EXPECT().
			Logf(ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == types.AuditActionNaisNamespaceCreateNamespace
			}), mock.Anything, team.Slug, environment).
			Return().
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		assert.NoError(t, r.Reconcile(ctx, input))

		msgs := srv.Messages()
		assert.Len(t, msgs, 1)

		publishRequest := &nais_namespace_reconciler.NaisdRequest{}
		json.Unmarshal(msgs[0].Data, publishRequest)

		createNamespaceRequest := &nais_namespace_reconciler.NaisdCreateNamespace{}
		json.Unmarshal(publishRequest.Data, createNamespaceRequest)

		assert.Equal(t, teamSlug, createNamespaceRequest.Name)
		assert.Equal(t, teamProjectID, createNamespaceRequest.GcpProject)
		assert.Equal(t, googleWorkspaceEmail, createNamespaceRequest.GroupEmail)
		assert.Equal(t, cnrmEmail, createNamespaceRequest.CNRMEmail)
		assert.Equal(t, "#env-channel", createNamespaceRequest.SlackAlertsChannel)
		assert.Equal(t, azureGroupID.String(), createNamespaceRequest.AzureGroupID)
	})

	t.Run("delete namespaces", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-"+environment)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
			Once()

		// var theState *reconcilers.NaisNamespaceState
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.NaisNamespaceState)
				state.Namespaces[environment] = team.Slug
				// theState = state
			}).
			Return(nil).
			Once()
		// database.
		// 	On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, theState).
		// 	Return(nil).
		// 	Once()
		database.
			On("RemoveReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.EXPECT().
			Logf(ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == types.AuditActionNaisNamespaceDeleteNamespace
			}), mock.Anything, team.Slug, environment).
			Return().
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		assert.NoError(t, r.Delete(ctx, team.Slug, input.CorrelationID))

		msgs := srv.Messages()
		assert.Len(t, msgs, 1)

		publishRequest := &nais_namespace_reconciler.NaisdRequest{}
		json.Unmarshal(msgs[0].Data, publishRequest)

		deleteNamespaceRequest := &nais_namespace_reconciler.NaisdDeleteNamespace{}
		json.Unmarshal(publishRequest.Data, deleteNamespaceRequest)

		assert.Equal(t, teamSlug, deleteNamespaceRequest.Name)
	})

	t.Run("create legacy namespaces", func(t *testing.T) {
		var (
			teamSlug         = "slug"
			environment      = "dev-gcp"
			clusterProjectID = "nais-dev-2e7b"
		)

		clusters := gcp.Clusters{
			environment: gcp.Cluster{
				TeamsFolderID: 123,
				ProjectID:     clusterProjectID,
			},
		}
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-"+environment)
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
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
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.NaisNamespaceState) bool {
				return state.Namespaces[environment] == team.Slug
			})).
			Return(nil).
			Once()
		database.
			On("GetSlackAlertsChannels", ctx, team.Slug).
			Return(map[string]string{}, nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.EXPECT().
			Logf(ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == types.AuditActionNaisNamespaceCreateNamespace
			}), mock.Anything, team.Slug, environment).
			Return().
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
		assert.NoError(t, r.Reconcile(ctx, input))

		msgs := srv.Messages()
		assert.Len(t, msgs, 1)

		publishRequest := &nais_namespace_reconciler.NaisdRequest{}
		json.Unmarshal(msgs[0].Data, publishRequest)

		createNamespaceRequest := &nais_namespace_reconciler.NaisdCreateNamespace{}
		json.Unmarshal(publishRequest.Data, createNamespaceRequest)

		assert.Equal(t, teamSlug, createNamespaceRequest.Name)
		assert.Equal(t, teamProjectID, createNamespaceRequest.GcpProject)
		assert.Equal(t, googleWorkspaceEmail, createNamespaceRequest.GroupEmail)
		assert.Equal(t, cnrmEmail, createNamespaceRequest.CNRMEmail)
		assert.Equal(t, azureGroupID.String(), createNamespaceRequest.AzureGroupID)
		assert.Equal(t, slackChannel, createNamespaceRequest.SlackAlertsChannel)
	})

	t.Run("create namespaces with additional legacy mappings", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-dev", "naisd-console-prod-fss")
		defer close()

		const onpremCluster = "prod-fss"
		onpremClusters := []string{onpremCluster}

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
		log.
			On("WithTeamSlug", teamSlug).
			Return(log).
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
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.NaisNamespaceState) bool {
				return state.Namespaces[environment] == team.Slug && state.Namespaces[onpremCluster] == team.Slug
			})).
			Return(nil).
			Once()
		database.
			On("GetSlackAlertsChannels", ctx, team.Slug).
			Return(map[string]string{}, nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.EXPECT().
			Logf(ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == types.AuditActionNaisNamespaceCreateNamespace
			}), mock.Anything, team.Slug, environment).
			Return().
			Once()

		auditLogger.EXPECT().
			Logf(ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Type == "team" && targets[0].Identifier == string(team.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == input.CorrelationID && fields.Action == types.AuditActionNaisNamespaceCreateNamespace
			}), mock.Anything, team.Slug, onpremCluster).
			Return().
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, clusters, domain, managementProjectID, azureEnabled, pubsubClient, log, onpremClusters)
		assert.NoError(t, r.Reconcile(ctx, input))

		msgs := srv.Messages()
		assert.Len(t, msgs, 2)
		onpremNamespaceCreated := false
		gcpNamespaceCreated := false
		for _, msg := range msgs {
			publishRequest := &nais_namespace_reconciler.NaisdRequest{}
			json.Unmarshal(msg.Data, publishRequest)

			createNamespaceRequest := &nais_namespace_reconciler.NaisdCreateNamespace{}
			json.Unmarshal(publishRequest.Data, createNamespaceRequest)

			if createNamespaceRequest.GcpProject == "" {
				assert.Equal(t, "", createNamespaceRequest.CNRMEmail)
				onpremNamespaceCreated = true
			} else {
				assert.Equal(t, teamProjectID, createNamespaceRequest.GcpProject)
				assert.Equal(t, cnrmEmail, createNamespaceRequest.CNRMEmail)
				gcpNamespaceCreated = true
			}

			assert.Equal(t, teamSlug, createNamespaceRequest.Name)
			assert.Equal(t, googleWorkspaceEmail, createNamespaceRequest.GroupEmail)
			assert.Equal(t, azureGroupID.String(), createNamespaceRequest.AzureGroupID)
			assert.Equal(t, slackChannel, createNamespaceRequest.SlackAlertsChannel)
		}

		assert.True(t, onpremNamespaceCreated)
		assert.True(t, gcpNamespaceCreated)
	})

	t.Run("environment in state no longer active", func(t *testing.T) {
		srv, pubsubClient, close := getPubsubServerAndClient(ctx, managementProjectID, "naisd-console-dev")
		defer close()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameNaisNamespace).
			Return(log).
			Once()
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
			On("SetReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.NaisNamespaceState) bool {
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
		database.
			On("GetSlackAlertsChannels", ctx, team.Slug).
			Return(map[string]string{}, nil).
			Once()

		r := nais_namespace_reconciler.New(database, auditLogger, gcp.Clusters{}, domain, managementProjectID, azureEnabled, pubsubClient, log, noOnpremClusters)
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

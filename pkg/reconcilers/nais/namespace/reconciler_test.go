package nais_namespace_reconciler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_token_source"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconcile(t *testing.T) {
	ctx := context.Background()
	team := db.Team{Team: &sqlc.Team{Slug: "slug"}}
	input := reconcilers.Input{
		CorrelationID: uuid.New(),
		Team:          team,
	}
	domain := "example.com"
	projectID := "some-project-123"
	azureEnabled := true

	auditLogger := auditlogger.NewMockAuditLogger(t)
	logger := logger.NewMockLogger(t)
	tokenSource, _ := google_token_source.NewFromConfig(&config.Config{
		GoogleManagementProjectID: projectID,
		TenantDomain:              domain,
	}).GCP(ctx)

	t.Run("unable to load namespace state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameNaisNamespace, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, domain, projectID, azureEnabled, tokenSource, logger)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load NAIS namespace state for team "slug"`)
	})

	t.Run("unable to load GCP project state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, domain, projectID, azureEnabled, tokenSource, logger)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `unable to load GCP project state for team "slug"`)
	})

	t.Run("no GCP projects in state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, domain, projectID, azureEnabled, tokenSource, logger)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no GCP project state exists for team "slug"`)
	})

	t.Run("unable to get google group email", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects["dev"] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: "some-project-id",
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, domain, projectID, azureEnabled, tokenSource, logger)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no workspace admin state exists for team "slug"`)
	})

	t.Run("unable to get azure group id", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, nais_namespace_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGcpProjectState)
				state.Projects["dev"] = reconcilers.GoogleGcpEnvironmentProject{
					ProjectID: "some-project-id",
				}
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				email := "group-email@example.com"
				state.GroupEmail = &email
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		r := nais_namespace_reconciler.New(database, auditLogger, domain, projectID, azureEnabled, tokenSource, logger)
		err := r.Reconcile(ctx, input)
		assert.ErrorContains(t, err, `no Azure state exists for team "slug"`)
	})
}

package google_gcp_reconciler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconcile(t *testing.T) {
	const (
		env              = "prod"
		teamFolderID     = 123
		clusterProjectID = "some-project-123"
		tenantName       = "example"
		tenantDomain     = "example.com"
		cnrmRoleName     = "organizations/123/roles/name"
		billingAccount   = "billingAccounts/123"
	)

	clusters := google_gcp_reconciler.ClusterInfo{
		env: {
			TeamFolderID: teamFolderID,
			ProjectID:    clusterProjectID,
		},
	}

	teamID := uuid.New()
	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{ID: teamID}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	t.Run("fail early when unable to load reconciler state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.ID, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when unable to load google workspace state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.ID, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when google workspace state is missing group email", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "no Google Workspace group exists")
	})

	t.Run("no error when we have no clusters", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.ID, mock.Anything).
			Run(func(args mock.Arguments) {
				email := "mail@example.com"
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &email
			}).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, google_gcp_reconciler.ClusterInfo{}, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount)

		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}

func TestGenerateProjectID(t *testing.T) {
	// different organization names don't show up in name, but are reflected in the hash
	assert.Equal(t, "happyteam-prod-488a", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "happyteam"))
	assert.Equal(t, "happyteam-prod-5534", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam"))

	// environments that get truncated produce different hashes
	assert.Equal(t, "sadteam-prod-04d4", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "sadteam"))
	assert.Equal(t, "sadteam-prod-6ce6", google_gcp_reconciler.GenerateProjectID("nais.io", "producers", "sadteam"))

	// team names that get truncated produce different hashes
	assert.Equal(t, "happyteam-is-very-ha-prod-4b2d", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy"))
	assert.Equal(t, "happyteam-is-very-ha-prod-4801", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy-and-altogether-too-long"))
}

func TestGetClusterInfoFromJson(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		info, err := google_gcp_reconciler.GetClusterInfoFromJson("")
		assert.Nil(t, info)
		assert.EqualError(t, err, "parse GCP cluster info: EOF")
	})

	t.Run("empty JSON object", func(t *testing.T) {
		info, err := google_gcp_reconciler.GetClusterInfoFromJson("{}")
		assert.NoError(t, err)
		assert.Empty(t, info)
	})

	t.Run("JSON with clusters", func(t *testing.T) {
		jsonData := `{
			"env1": {"teams_folder_id": "123", "project_id": "some-id-123"},
			"env2": {"teams_folder_id": "456", "project_id": "some-id-456"}
		}`
		info, err := google_gcp_reconciler.GetClusterInfoFromJson(jsonData)
		assert.NoError(t, err)

		assert.Contains(t, info, "env1")
		assert.Equal(t, int64(123), info["env1"].TeamFolderID)
		assert.Equal(t, "some-id-123", info["env1"].ProjectID)

		assert.Contains(t, info, "env2")
		assert.Equal(t, int64(456), info["env2"].TeamFolderID)
		assert.Equal(t, "some-id-456", info["env2"].ProjectID)
	})
}

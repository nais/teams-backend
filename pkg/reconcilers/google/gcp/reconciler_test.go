package google_gcp_reconciler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	const (
		env              = "prod"
		teamFolderID     = 123
		clusterProjectID = "some-project-123"
		tenantName       = "example"
		tenantDomain     = "example.com"
		cnrmRoleName     = "organizations/123/roles/name"
		billingAccount   = "billingAccounts/123"
	)

	clusters := gcp.Clusters{
		env: {
			TeamsFolderID: teamFolderID,
			ProjectID:     clusterProjectID,
		},
	}

	emptyMap := make(map[string]string, 0)

	teamSlug := slug.Slug("slug")
	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: teamSlug}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	t.Run("fail early when unable to load reconciler state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, emptyMap, log)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when unable to load google workspace state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, emptyMap, log)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when google workspace state is missing group email", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, emptyMap, log)

		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "no Google Workspace group exists")
	})

	t.Run("no error when we have no clusters", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				email := "mail@example.com"
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &email
			}).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}
		reconciler := google_gcp_reconciler.New(database, auditLogger, gcp.Clusters{}, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, emptyMap, log)

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

	// project id with double hyphens
	assert.Equal(t, "hapyteam-is-very-ha-prod-fd5d", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "hapyteam-is-very-ha-a"))

	// environment with hyphen as 4th character in environment
	assert.Equal(t, "hapyteam-is-happy-pro-2a15", google_gcp_reconciler.GenerateProjectID("bais.io", "pro-duction", "hapyteam-is-happy"))
}

func TestGetProjectDisplayName(t *testing.T) {
	tests := []struct {
		slug        string
		environment string
		displayName string
	}{
		{"some-slug", "prod", "some-slug-prod"},
		{"some-slug", "production", "some-slug-production"},
		{"some-verry-unnecessarily-long-slug", "dev", "some-verry-unnecessarily-l-dev"},
		{"some-verry-unnecessarily-long-slug", "prod", "some-verry-unnecessarily-prod"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.displayName, google_gcp_reconciler.GetProjectDisplayName(slug.Slug(tt.slug), tt.environment))
	}
}

func TestCnrmServiceAccountNameAndAccountID(t *testing.T) {
	tests := []struct {
		slug               slug.Slug
		projectID          string
		generatedName      string
		generatedAccountID string
	}{
		{
			"some-slug",
			"foo-bar-123",
			"projects/foo-bar-123/serviceAccounts/cnrm-some-slug-ba7f@foo-bar-123.iam.gserviceaccount.com",
			"cnrm-some-slug-ba7f",
		},
		{
			"slug",
			"foobar-barfoo-123a",
			"projects/foobar-barfoo-123a/serviceAccounts/cnrm-slug-cd03@foobar-barfoo-123a.iam.gserviceaccount.com",
			"cnrm-slug-cd03",
		},
		{
			"some-team-slug-that-is-waaaaaaaaaaay-to-long",
			"foo-bar-123",
			"projects/foo-bar-123/serviceAccounts/cnrm-some-team-slug-that-9233@foo-bar-123.iam.gserviceaccount.com",
			"cnrm-some-team-slug-that-9233",
		},
		{
			"someteam-slug-that-is-waaaaaaaaaaay-to-long",
			"foo-bar-123",
			"projects/foo-bar-123/serviceAccounts/cnrm-someteam-slug-that-i-d830@foo-bar-123.iam.gserviceaccount.com",
			"cnrm-someteam-slug-that-i-d830",
		},
	}
	for _, tt := range tests {
		name, accountID := google_gcp_reconciler.CnrmServiceAccountNameAndAccountID(tt.slug, tt.projectID)
		assert.Equal(t, tt.generatedName, name)
		assert.Equal(t, tt.generatedAccountID, accountID)
	}
}

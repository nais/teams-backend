//go:build adhoc_integration_test

package nais_namespace_reconciler_test

import (
	"context"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/test"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/stretchr/testify/assert"
)

// FIXME: Test is currently failing
func TestNaisNamespaceReconciler(t *testing.T) {
	ctx := context.Background()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	db := test.GetTestDB()

	rec, err := nais_namespace_reconciler.NewFromConfig(db, cfg, logger)
	if err != nil {
		panic(err)
	}

	sysid := uuid.New()

	teamName := dbmodels.Slug("foo")
	err = rec.Reconcile(ctx, reconcilers.Input{
		System: &dbmodels.System{
			Model: dbmodels.Model{
				ID: &sysid,
			},
		},
		Team: &dbmodels.Team{
			Slug: &teamName,
			SystemState: []*dbmodels.SystemState{
				{
					SystemID:    &sysid,
					Environment: helpers.Strp("dev"),
					Key:         "google-project-id",
					Value:       "this-is-the-google-project-id",
				},
			},
		},
	})

	assert.NoError(t, err)
}

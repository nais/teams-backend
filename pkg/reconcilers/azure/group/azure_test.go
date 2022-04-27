//go:build adhoc_integration_test

package azure_group_reconciler_test

import (
	"context"
	"testing"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const teamName = "myteam"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

	creds := clientcredentials.Config{}
	reconciler := azure_group.New(logger, creds)

	err := reconciler.Reconcile(ctx, reconcilers.Input{
		System:          nil,
		Synchronization: nil,
		Team: &dbmodels.Team{
			Slug:    strp(teamName),
			Name:    strp(teamName),
			Purpose: strp(teamName),
		},
	})

	assert.NoError(t, err)
}

func strp(s string) *string {
	return &s
}

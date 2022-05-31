//go:build adhoc_integration_test

package azure_group_reconciler_test

import (
	"context"
	helpers "github.com/nais/console/pkg/console"
	"sync"
	"testing"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
)

func TestReconcile(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	cfg.Azure.Enabled = true

	ch := make(chan *dbmodels.AuditLog, 2048)
	logger := auditlogger.New(ch)
	rec, err := azure_group.NewFromConfig(cfg, logger)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := reconcilers.Input{
		System:          nil,
		Synchronization: nil,
		Team: &dbmodels.Team{
			Slug:    helpers.Strp("foobarbaz"),
			Name:    helpers.Strp("test group, can be deleted"),
			Purpose: helpers.Strp("this is just a test"),
			Users:   []*dbmodels.User{},
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for logline := range ch {
			t.Log(logline.Message)
		}
		wg.Done()
	}()

	err = rec.Reconcile(ctx, input)
	close(ch)

	assert.NoError(t, err)

	wg.Wait()
}

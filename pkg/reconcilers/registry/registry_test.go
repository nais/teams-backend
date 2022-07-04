package registry_test

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func reconciler() registry.ReconcilerFactory {
	return func(*gorm.DB, *config.Config, dbmodels.System, auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
		return nil, nil
	}
}

func TestRegister(t *testing.T) {
	rec1 := reconciler()
	rec2 := reconciler()
	rec3 := reconciler()

	assert.Empty(t, registry.Reconcilers())
	registry.Register("rec1", rec1)
	registry.Register("rec2", rec2)
	registry.Register("rec2", rec2) // Same name as previous
	registry.Register("rec3", rec3)

	reconcilers := registry.Reconcilers()
	assert.Len(t, reconcilers, 3)
	assert.Equal(t, "rec1", reconcilers[0].Name)
	assert.Equal(t, "rec2", reconcilers[1].Name)
	assert.Equal(t, "rec3", reconcilers[2].Name)
}
